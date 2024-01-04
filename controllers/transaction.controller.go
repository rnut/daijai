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
	var transaction []models.Transaction
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

// / get transactions by PONumber
func (mc *TransactionController) GetTransactionsByPONumber(c *gin.Context) {
	ponumber := c.Param("id")
	var transaction []models.Transaction
	if err := mc.
		DB.
		Preload("Inventory").
		Preload("Material").
		Preload("Receipt").
		Where("po_number = ?", ponumber).
		Find(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transaction"})
		return
	}
	c.JSON(http.StatusOK, transaction)
}
