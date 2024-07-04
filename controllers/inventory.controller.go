package controllers

import (
	"daijai/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InventoryController struct {
	DB *gorm.DB
	BaseController
}

func NewInventoryController(db *gorm.DB) *InventoryController {
	return &InventoryController{
		DB: db,
	}
}

func (mc *InventoryController) CreateInventory(c *gin.Context) {
	var request models.Inventory
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a new Project
	if err := mc.DB.Create(&request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create inventory"})
		return
	}
	c.JSON(http.StatusCreated, request)
}

func (mc *InventoryController) UpdateInventory(c *gin.Context) {
	id := c.Param("id")
	var request models.Inventory
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var inventory models.Inventory
	if err := mc.DB.First(&inventory, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
		return
	}

	if err := mc.DB.Model(&inventory).Updates(&request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update inventory"})
		return
	}
	c.JSON(http.StatusOK, inventory)
}

func (mc *InventoryController) DeleteInventory(c *gin.Context) {
	id := c.Param("id")
	var inventory models.Inventory
	if err := mc.DB.First(&inventory, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
		return
	}

	if err := mc.DB.Delete(&inventory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete inventory"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Inventory deleted successfully"})
}

// get all inventory
func (mc *InventoryController) GetInventories(c *gin.Context) {
	var inventory []models.Inventory
	if err := mc.DB.Find(&inventory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inventory"})
		return
	}
	c.JSON(http.StatusOK, inventory)
}

// get inventory by slug
func (mc *InventoryController) GetInventoryBySlug(c *gin.Context) {
	slug := c.Param("slug")
	var response struct {
		Inventory  models.Inventory
		Categories []models.Category
	}
	var inventory models.Inventory
	if err := mc.
		DB.
		First(&inventory, "slug = ?", slug).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inventory"})
		return
	} else if inventory.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
		return
	}
	response.Inventory = inventory

	// get categories
	var categories []models.Category
	if err := mc.DB.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get categories"})
		return
	}
	response.Categories = categories
	c.JSON(http.StatusOK, response)
}

// // get InventoryTransaction by id
// func (mc *InventoryController) GetInventoryTransaction(c *gin.Context) {
// 	id := c.Param("id")
// 	var inventory []models.InventoryTransaction
// 	if err := mc.DB.Where("inventory_id = ?", id).Find(&inventory).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inventory"})
// 		return
// 	} else if len(inventory) == 0 {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
// 		return
// 	}
// }

// // get InventoryTransaction by inventory id and material id
// func (mc *InventoryController) GetInventoryTransactionByMaterial(c *gin.Context) {
// 	inventoryID := c.Param("inventory_id")
// 	materialID := c.Param("material_id")
// 	var inventory []models.InventoryTransaction
// 	if err := mc.DB.Where("inventory_id = ? AND material_id = ?", inventoryID, materialID).Find(&inventory).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inventory"})
// 		return
// 	} else if len(inventory) == 0 {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
// 		return
// 	}
// }

// // get all inventory transaction
// func (mc *InventoryController) GetAllInventoryTransactions(c *gin.Context) {
// 	var inventory []models.InventoryTransaction
// 	if err := mc.DB.Find(&inventory).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inventory"})
// 		return
// 	}
// 	c.JSON(http.StatusOK, inventory)
// }

// / transfer material from one inventory to another
func (mc *InventoryController) TransferMaterial(c *gin.Context) {

	var request struct {
		FromInventoryID uint  `json:"fromInventoryID"`
		ToInventoryID   uint  `json:"toInventoryID"`
		MaterialID      uint  `json:"materialID"`
		Quantity        int64 `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// get from inventory
	var fromInventory models.Inventory
	if err := mc.DB.First(&fromInventory, request.FromInventoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
		return
	}

	// get to inventory
	var toInventory models.Inventory
	if err := mc.DB.First(&toInventory, request.ToInventoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
		return
	}

	// get material
	var material models.Material
	if err := mc.DB.First(&material, request.MaterialID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material not found"})
		return
	}

	// get inventory material
	var fromInventoryMaterials []models.InventoryMaterial
	if err := mc.DB.
		Where("inventory_id = ?", fromInventory.ID).
		Where("material_id = ?", material.ID).
		Where("is_out_of_stock = ?", false).
		Where("available_qty != ?", 0).
		Find(&fromInventoryMaterials).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory material not found"})
		return
	}

	// get user
	var uid uint
	if err := mc.GetUserID(c, &uid); err != nil {
		mc.LogErrorAndSendBadRequest(c, err.Error())
		return
	}
	var member models.Member
	if err := mc.getUserDataByUserID(mc.DB, uid, &member); err != nil {
		mc.LogErrorAndSendBadRequest(c, err.Error())
		return
	}

	if err := mc.DB.Transaction(func(tx *gorm.DB) error {
		// create TransferMaterial
		transferMaterial := models.TransferMaterial{
			FromInventoryID: fromInventory.ID,
			ToInventoryID:   toInventory.ID,
			MaterialID:      material.ID,
			Quantity:        request.Quantity,
			CreatedByID:     member.ID,
		}
		if err := tx.Create(&transferMaterial).Error; err != nil {
			return err
		}

		// loop through the ifromInventoryMaterials
		targetQty := request.Quantity
		for _, iv := range fromInventoryMaterials {
			// if targetQty is 0, break the loop
			if targetQty == 0 {
				break
			}

			// calculate the transferQty
			var transferQty int64
			if iv.AvailableQty >= targetQty {
				transferQty = targetQty
			} else {
				transferQty = iv.AvailableQty
			}

			// create a new inventory material
			newInventoryMaterial := models.InventoryMaterial{
				InventoryID:           toInventory.ID,
				MaterialID:            material.ID,
				AdjustmentID:          nil,
				Quantity:              transferQty,
				AvailableQty:          transferQty,
				IsOutOfStock:          false,
				Price:                 iv.Price,
				TransferMaterialID:    &transferMaterial.ID,
				InventoryMaterialType: models.InventoryMaterialType_Transfer,
				ReceiptID:             iv.ReceiptID,
			}
			if err := tx.Create(&newInventoryMaterial).Error; err != nil {
				return err
			}

			// calculate the updatedQuantityFromInventory
			updatedQuantityFromInventory := iv.Quantity - transferQty
			updatedFromInventoryAvailableQty := iv.AvailableQty - transferQty

			// create transfer-out transaction
			transferOutTransaction := models.InventoryMaterialTransaction{
				InventoryMaterialID:      iv.ID,
				Quantity:                 transferQty,
				InventoryType:            models.InventoryType_TRANSFER,
				InventoryTypeDescription: models.InventoryTypeDescription_TRANSFER_OUT,
				ExistingQuantity:         iv.Quantity,
				ExistingReserve:          iv.Reserve,
				UpdatedQuantity:          updatedQuantityFromInventory,
				UpdatedReserve:           iv.Reserve,
				ReceiptID:                iv.ReceiptID,
				TransferMaterialID:       &transferMaterial.ID,
			}
			if err := tx.Create(&transferOutTransaction).Error; err != nil {
				return err
			}

			// create transfer-in transaction
			transferInTransaction := models.InventoryMaterialTransaction{
				InventoryMaterialID:      newInventoryMaterial.ID,
				Quantity:                 transferQty,
				InventoryType:            models.InventoryType_TRANSFER,
				InventoryTypeDescription: models.InventoryTypeDescription_TRANSFER_IN,
				ExistingQuantity:         0,
				ExistingReserve:          0,
				UpdatedQuantity:          transferQty,
				UpdatedReserve:           0,
				ReceiptID:                iv.ReceiptID,
				TransferMaterialID:       &transferMaterial.ID,
			}
			if err := tx.Create(&transferInTransaction).Error; err != nil {
				return err
			}

			// update the fromInventoryMaterial
			if updatedFromInventoryAvailableQty == 0 {
				iv.IsOutOfStock = true
			}
			if err := tx.Model(&iv).Updates(models.InventoryMaterial{
				Quantity:     updatedQuantityFromInventory,
				AvailableQty: updatedFromInventoryAvailableQty,
				IsOutOfStock: iv.IsOutOfStock,
			}).Error; err != nil {
				return err
			}

			// update targetQty
			targetQty -= transferQty
		}
		mc.SumMaterial(tx, "transfer", material.ID, fromInventory.ID)
		mc.SumMaterial(tx, "transfer", material.ID, toInventory.ID)

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to transfer material", "detail": err.Error()})
		return
	}
	if err := mc.DB.
		Preload("Sums", "inventory_id = ?", fromInventory.ID).
		First(&material, request.MaterialID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get material"})
	}
	c.JSON(http.StatusOK, material)
}

// / calculate cost of transfer material
func (mc *InventoryController) CalculateCostOfTransferMaterial(c *gin.Context) {
	var request struct {
		FromInventoryID uint `json:"fromInventoryID"`
		MaterialID      uint `json:"materialID"`
		Quantity        int  `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// get from inventory
	var fromInventory models.Inventory
	if err := mc.DB.First(&fromInventory, request.FromInventoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
		return
	}

	// get material
	var material models.Material
	if err := mc.DB.First(&material, request.MaterialID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material not found"})
		return
	}

	// get inventory material
	var fromInventoryMaterials []models.InventoryMaterial
	if err := mc.DB.
		Where("inventory_id = ?", fromInventory.ID).
		Where("material_id = ?", material.ID).
		Where("is_out_of_stock = ?", false).
		Where("available_qty != ?", 0).
		Find(&fromInventoryMaterials).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory material not found"})
		return
	}

	var totalCost int64
	var maxTransferQty int64
	for _, iv := range fromInventoryMaterials {
		// calculate the transferQty
		var transferQty int64
		if iv.AvailableQty >= int64(request.Quantity) {
			transferQty = int64(request.Quantity)
		} else {
			transferQty = iv.AvailableQty
		}
		totalCost += (transferQty * iv.Price) / 100
		maxTransferQty += transferQty
	}
	c.JSON(http.StatusOK, gin.H{"totalCost": totalCost, "maxTransferQty": maxTransferQty})
}
