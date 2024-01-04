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
		for _, v := range receipt.ReceiptMaterials {
			v.IsApproved = true
			if err := tx.Save(&v).Error; err != nil {
				return err
			}
			// get last InventoryTransaction
			var lastTransaction models.Transaction
			if err := tx.Last(&lastTransaction, "material_id = ? AND inventory_id = ?", v.MaterialID, receipt.InventoryID).Error; err != nil {
				lastTransaction.TotalQuantity = 0
				lastTransaction.TotalReserve = 0
			}

			t := models.Transaction{
				MaterialID:     v.MaterialID,
				InventoryID:    receipt.InventoryID,
				Quantity:       lastTransaction.TotalQuantity,
				Reserve:        lastTransaction.TotalReserve,
				QuantityChange: v.Quantity,
				ReserveChange:  0,
				TotalQuantity:  lastTransaction.TotalQuantity + v.Quantity,
				TotalReserve:   lastTransaction.TotalReserve,
				Price:          v.Price,
				Type:           "receipt",
				Ref:            receipt.Slug,
				PONumber:       receipt.PONumber,
				ReceiptID:      &receipt.ID,
			}
			if err := tx.Create(&t).Error; err != nil {
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
			if err := tx.Save(&inventoryMaterial).Error; err != nil {
				return err
			}
		}

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
