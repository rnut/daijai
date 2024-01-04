package controllers

import (
	"daijai/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InventoryController struct {
	DB *gorm.DB
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
	c.JSON(http.StatusCreated, gin.H{"message": "Inventory created successfully"})
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

// get inventory by id
func (mc *InventoryController) GetInventoryByID(c *gin.Context) {
	id := c.Param("id")
	var inventory models.Inventory
	if err := mc.
		DB.
		Preload("InventoryMaterials").
		First(&inventory, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inventory"})
		return
	} else if inventory.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
		return
	}
	c.JSON(http.StatusOK, inventory)
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
