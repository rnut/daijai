package controllers

import (
	"daijai/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TransactionController struct {
	DB *gorm.DB
}

func NewTransactionController(db *gorm.DB) *TransactionController {
	return &TransactionController{
		DB: db,
	}
}

func (mc *TransactionController) GetTransactions(c *gin.Context) {
	var transaction []models.AppLog
	if err := mc.
		DB.
		Preload("Inventory").
		Preload("Material").
		Find(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transaction"})
		return
	}
	c.JSON(http.StatusOK, transaction)
}

func (mc *TransactionController) GetTransactionsGroupByInventory(c *gin.Context) {
	var inventories []models.Inventory
	if err := mc.
		DB.
		Preload("Transactions").
		Preload("Transactions.Material").
		Find(&inventories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transaction"})
		return
	}
	c.JSON(http.StatusOK, inventories)
}

// / get inventory material transactions by receipt.PONumber
func (mc *TransactionController) GetInventoryMaterialTransactionsByPONumber(c *gin.Context) {
	ponumber := c.Param("poNumber")
	var ivtMats []models.InventoryMaterial
	if err := mc.
		DB.
		Joins("Receipt").
		Where("po_number = ?", ponumber).
		Find(&ivtMats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inventory material transactions"})
		return
	}

	var ivtIDs []uint
	for _, ivtMat := range ivtMats {
		ivtIDs = append(ivtIDs, ivtMat.ID)
	}

	var transactions []models.InventoryMaterialTransaction
	if err := mc.
		DB.
		Where("inventory_material_id IN ?", ivtIDs).
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inventory material transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions, "ivtIDs": ivtIDs})
}
