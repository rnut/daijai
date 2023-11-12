package controllers

import (
	"daijai/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

// WithdrawalController handles CRUD operations for Withdrawal model.
type WithdrawalController struct {
	DB *gorm.DB
}

// NewWithdrawalController creates a new instance of WithdrawalController.
func NewWithdrawalController(db *gorm.DB) *WithdrawalController {
	return &WithdrawalController{
		DB: db,
	}
}

func (wc *WithdrawalController) GetAllWithdrawals(c *gin.Context) {
	var withdrawals []models.Withdrawal

	if err := wc.DB.Preload("WithdrawalMaterials").Preload("WithdrawalMaterials.Material").Find(&withdrawals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve withdrawals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"withdrawals": withdrawals})
}

// CreateWithdrawal handles the creation of a new withdrawal transaction.
func (wc *WithdrawalController) CreateWithdrawal(c *gin.Context) {
	var request struct {
		Withdrawal          models.Withdrawal
		WithdrawalMaterials []models.WithdrawalMaterial
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := wc.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&request.Withdrawal).Error; err != nil {
			return err
		}

		// Associate the bombs with the drawing
		for i := range request.WithdrawalMaterials {
			request.WithdrawalMaterials[i].WithdrawalID = request.Withdrawal.ID
			if err := tx.Create(&request.WithdrawalMaterials[i]).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create drawing"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Withdrawal created successfully", "withdrawal": request.Withdrawal})
}

// // ApproveWithdrawal approves a withdrawal transaction and updates the material quantity.
func (wc *WithdrawalController) ApproveWithdrawal(c *gin.Context) {
	withdrawalID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid withdrawal ID"})
		return
	}

	var withdrawal models.Withdrawal
	if err := wc.DB.Preload("WithdrawalMaterials").Preload("WithdrawalMaterials.Material").First(&withdrawal, withdrawalID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal not found", "id": withdrawalID})
		return
	}

	// Check if the withdrawal is already approved
	if withdrawal.IsApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Withdrawal is already approved"})
		return
	}

	tx := wc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve withdrawal"})
	}

	// Update the associated materials' quantity
	for i, withdrawalMaterial := range withdrawal.WithdrawalMaterials {
		material := models.Material{}
		if err := wc.DB.First(&material, withdrawalMaterial.MaterialID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve material"})
			return
		}

		// Update the material's quantity
		if material.InUseQuantity >= withdrawalMaterial.Quantity {
			material.InUseQuantity -= withdrawalMaterial.Quantity
		} else {
			q := withdrawalMaterial.Quantity - material.InUseQuantity
			material.InUseQuantity = 0
			material.Quantity -= q
		}

		if material.Quantity < 0 {
			// Handle insufficient quantity error
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient quantity"})
			return
		}

		if err := tx.Save(&material).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update material"})
			return
		}

		// set material back to withdrawalMaterials
		withdrawal.WithdrawalMaterials[i].Material = material
	}

	// Mark the withdrawal as approved
	withdrawal.IsApproved = true
	if err := tx.Save(&withdrawal).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save approval withdrawal"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit withdrawal"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal approved successfully", "withdrawal": withdrawal})
}
