package controllers

import (
	"daijai/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReceiptController struct {
	DB *gorm.DB
	BaseController
}

func NewReceipt(db *gorm.DB) *ReceiptController {
	return &ReceiptController{
		DB: db,
	}
}

// CreateReceipt creates a new Receipt entry.
func (rc *ReceiptController) CreateReceipt(c *gin.Context) {
	var uid uint
	if err := rc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := rc.getUserDataByUserID(rc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var request models.Receipt
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request.CreatedByID = member.ID
	if err := rc.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&request).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Receipt"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Receipt created successfully", "receipt": request})
}

func (rc *ReceiptController) ApproveReceipt(c *gin.Context) {
	var uid uint
	if err := rc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := rc.getUserDataByUserID(rc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	var receipt models.Receipt
	if err := rc.
		DB.
		Preload("ReceiptMaterials").
		Preload("CreatedBy").
		First(&receipt, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}
	if err := rc.DB.Transaction(func(tx *gorm.DB) error {
		var matIDs []uint 
		
		matQuantity := make(map[uint]int64)
		matInventoryId := make(map[uint]uint)
		for _, v := range receipt.ReceiptMaterials {
			matIDs = append(matIDs, v.MaterialID)
			// update receipt material
			v.IsApproved = true
			if err := tx.Save(&v).Error; err != nil {
				return err
			}

			// create inventory material
			var inventoryMaterial models.InventoryMaterial
			inventoryMaterial.MaterialID = v.MaterialID
			inventoryMaterial.InventoryID = receipt.InventoryID
			inventoryMaterial.ReceiptID = receipt.ID
			inventoryMaterial.Quantity = v.Quantity
			inventoryMaterial.Reserve = 0
			inventoryMaterial.AvailabelQty = v.Quantity
			inventoryMaterial.IsOutOfStock = false
			inventoryMaterial.Price = v.Price
			if err := tx.Save(&inventoryMaterial).Error; err != nil {
				return err
			}

			// assign hashmap
			matQuantity[v.MaterialID] = v.Quantity
			matInventoryId[v.MaterialID] = inventoryMaterial.ID
			
			// create inventory materail transaction
			inventoryMaterialTransaction := models.InventoryMaterialTransaction {
				InventoryMaterialID: inventoryMaterial.ID,
				Quantity: inventoryMaterial.Quantity,
				InventoryType: models.InventoryType_INCOMING,
				InventoryTypeDescription: models.InventoryTypeDescription_INCOMINGRECEIPT,
				ExistingQuantity: 0,
				ExistingReserve: 0,
				UpdatedQuantity: inventoryMaterial.Quantity,
				UpdatedReserve: 0,
				ReceiptID: &receipt.ID,
			}
			if err := tx.Save(&inventoryMaterialTransaction).Error; err != nil {
				return err
			}
		}

		// get waiting material order boms
		var orderBoms []models.OrderBom
		withdrawStatuses := []string{models.OrderStatus_Waiting, models.OrderStatus_InProgress}
		if err := rc.
			DB.
			Joins("Bom").
			Joins("Order").
			Where("is_full_filled = ?", false).
			Where("withdraw_status IN (?)", withdrawStatuses).
			Where("material_id IN ?", matIDs).
			Find(&orderBoms).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch OrderBoms"})
			return err
		}

		// loop order boms
		for _, orderBom := range orderBoms {
			target := orderBom.TargetQty - (orderBom.ReservedQty + orderBom.WithdrawedQty)
			matID := orderBom.Bom.MaterialID
			qouta := matQuantity[matID]

			if qouta == 0 {
				continue	
			}

			var quantity int64
			if qouta > target {
				quantity = target
			} else {
				quantity = qouta
			}

			inventoryMaterialID := matInventoryId[matID]

			// create order reserving
			orderReserving := models.OrderReserving {
				OrderID: orderBom.OrderID,
				OrderBomID: orderBom.ID,
				ReceiptID: receipt.ID,
				InventoryMaterialID: inventoryMaterialID,
				Quantity: quantity,
				Status: models.OrderReservingStatus_Reserved,
			}
			if err := tx.Save(&orderReserving).Error; err != nil {
				return err
			}

			// update order bom
			orderBom.ReservedQty += quantity
			if orderBom.ReservedQty == orderBom.TargetQty {
				orderBom.IsFullFilled = true
			}
			if err := tx.Save(&orderBom).Error; err != nil {
				return err
			}

			// update inventory material
			var inventoryMaterial models.InventoryMaterial
			if err := tx.First(&inventoryMaterial, inventoryMaterialID).Error; err != nil {
				return err
			}
			inventoryMaterial.Reserve += quantity
			inventoryMaterial.AvailabelQty -= quantity
			if inventoryMaterial.AvailabelQty == 0 {
				inventoryMaterial.IsOutOfStock = true
			}
			if err := tx.Save(&inventoryMaterial).Error; err != nil {
				return err
			}

			// create inventory material transaction
			ivmtReserve := models.InventoryMaterialTransaction {
				InventoryMaterialID: inventoryMaterialID,
				Quantity: quantity,
				InventoryType: models.InventoryType_RESERVE,
				InventoryTypeDescription: models.InventoryTypeDescription_FillFromReceipt,
				ExistingQuantity: qouta,
				ExistingReserve: 0,
				UpdatedQuantity: qouta,
				UpdatedReserve: quantity,
				OrderID: &orderBom.OrderID,
			}
			if err := tx.Save(&ivmtReserve).Error; err != nil {
				return err
			}

			// update order status to in-progress
			orderBom.Order.WithdrawStatus = models.OrderStatus_InProgress
			if err := tx.Save(&orderBom.Order).Error; err != nil {
				return err
			}

			matQuantity[matID] -= quantity
		}

		// update receipt status and approved by
		receipt.IsApproved = true
		receipt.ApprovedByID = &member.ID
		receipt.ApprovedBy = &member
		if err := tx.Save(&receipt).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve Receipt"})
		return
	}
	c.JSON(http.StatusOK, receipt)
}

func (rc *ReceiptController) GetAllReceipts(c *gin.Context) {
	var uid uint
	if err := rc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := rc.getUserDataByUserID(rc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var receipts []models.Receipt
	q := rc.DB.
		Preload("ReceiptMaterials.Material").
		Preload("CreatedBy").
		Preload("ApprovedBy")

	if member.Role == "admin" {
		q.Find(&receipts)
	} else {
		q.Find(&receipts, "created_by_id = ?", member.ID)
	}
	if err := q.Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Receipts"})
		return
	}

	c.JSON(http.StatusOK, receipts)
}

// GetReceipt gets a Receipt by ID.
func (rc *ReceiptController) GetReceipt(c *gin.Context) {
	id := c.Param("id")

	var receipt models.Receipt
	if err := rc.
		DB.
		Preload("ReceiptMaterials.Material.Category").
		Preload("CreatedBy").
		First(&receipt, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	c.JSON(http.StatusOK, receipt)
}

// GET Recript by slug
func (rc *ReceiptController) GetReceiptBySlug(c *gin.Context) {
	slug := c.Param("slug")

	var receipt models.Receipt
	if err := rc.
		DB.
		Preload("ReceiptMaterials.Material.Category").
		Preload("CreatedBy").
		Preload("ApprovedBy").
		First(&receipt, "slug = ?", slug).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	// get all inventory materials by receipt id
	var inventoryMaterials []models.InventoryMaterial
	if err := rc.
		DB.
		Preload("Material").
		Preload("Inventory").
		Preload("Transactions.Order.Drawing").
		Find(&inventoryMaterials, "receipt_id = ?", receipt.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"receipt": receipt, "inventoryMaterials": inventoryMaterials})
}

// UpdateReceipt updates a Receipt by ID.
func (rc *ReceiptController) UpdateReceipt(c *gin.Context) {
	id := c.Param("id")

	var request models.Receipt
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var receipt models.Receipt
	if err := rc.DB.First(&receipt, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	if err := rc.DB.Save(&receipt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Receipt"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Receipt updated successfully", "receipt": receipt})
}

// DeleteReceipt deletes a Receipt by ID.
func (rc *ReceiptController) DeleteReceipt(c *gin.Context) {
	id := c.Param("id")

	var receipt models.Receipt
	if err := rc.DB.First(&receipt, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	if err := rc.DB.Delete(&receipt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete Receipt"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Receipt deleted successfully"})
}
