package controllers

import (
	"daijai/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

// WithdrawalController handles CRUD operations for Withdrawal model.
type WithdrawalController struct {
	DB *gorm.DB
	BaseController
}

// NewWithdrawalController creates a new instance of WithdrawalController.
func NewWithdrawalController(db *gorm.DB) *WithdrawalController {
	return &WithdrawalController{
		DB: db,
	}
}

func (wc *WithdrawalController) GetWithdrawalByID(c *gin.Context) {
	objID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Withdrawal ID"})
		return
	}

	var withdrawal models.Withdrawal
	if err := wc.DB.
		Preload("Project").
		Preload("WithdrawalMaterials.Material.Category").
		Preload("CreatedBy").
		Preload("ApprovedBy").
		First(&withdrawal, objID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal not found"})
		return
	}

	c.JSON(http.StatusOK, withdrawal)
}

func (wc *WithdrawalController) GetAllWithdrawals(c *gin.Context) {
	var withdrawals []models.Withdrawal

	if err := wc.DB.
		Preload("Project").
		Preload("WithdrawalMaterials").
		Preload("WithdrawalMaterials.Material").
		Preload("WithdrawalMaterials.Material.Category").
		Preload("CreatedBy").
		Preload("ApprovedBy").
		Find(&withdrawals).Error; err != nil {
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

	var uid uint
	if err := wc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := wc.getUserDataByUserID(wc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := wc.DB.Transaction(func(tx *gorm.DB) error {
		request.Withdrawal.CreatedByID = member.ID
		if err := tx.Create(&request.Withdrawal).Error; err != nil {
			return err
		}

		for i := range request.WithdrawalMaterials {
			request.WithdrawalMaterials[i].WithdrawalID = request.Withdrawal.ID
			if err := tx.Create(&request.WithdrawalMaterials[i]).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Withdrawal created successfully", "withdrawal": request.Withdrawal})
}

func (wc *WithdrawalController) UpdateWithdrawal(c *gin.Context) {
	var request struct {
		Withdrawal struct {
			ProjectID uint
			Notes     string
		}
		WithdrawalMaterials []models.WithdrawalMaterial
	}

	withdrawalID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid drawing ID"})
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var withdrawal models.Withdrawal
	if err := wc.DB.Preload("Project").Preload("WithdrawalMaterials").Preload("WithdrawalMaterials.Material.Category").Preload("CreatedBy").First(&withdrawal, withdrawalID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal not found", "id": withdrawalID})
		return
	}

	if err := wc.DB.Transaction(func(tx *gorm.DB) error {
		pID := request.Withdrawal.ProjectID
		withdrawal.ProjectID = pID
		withdrawal.Project.ID = pID
		withdrawal.Notes = request.Withdrawal.Notes
		if err := tx.Save(&withdrawal).Error; err != nil {
			return err
		}

		// DELETE ALL WithdrawalMaterials
		for _, v := range withdrawal.WithdrawalMaterials {
			if err := tx.Delete(&models.WithdrawalMaterial{}, v.ID).Error; err != nil {
				return err
			}
		}

		for _, v := range request.WithdrawalMaterials {
			wm := models.WithdrawalMaterial{
				WithdrawalID: withdrawal.ID,
				MaterialID:   v.MaterialID,
				Quantity:     v.Quantity,
			}
			if err := tx.Create(&wm).Error; err != nil {
				return err
			}
			withdrawal.WithdrawalMaterials = append(withdrawal.WithdrawalMaterials, wm)
		}
		return nil
	}); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Update Withdraw successfully", "withdrawal": withdrawal})
}

// // ApproveWithdrawal approves a withdrawal transaction and updates the material quantity.
func (wc *WithdrawalController) ApproveWithdrawal(c *gin.Context) {
	var uid uint
	if err := wc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := wc.getUserDataByUserID(wc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if member.Role != "admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Permission Denied"})
		return
	}

	withdrawalID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid withdrawal ID"})
		return
	}

	var withdrawal models.Withdrawal
	if err := wc.DB.
		Preload("Project").
		Preload("WithdrawalMaterials.Material.Category").
		Preload("CreatedBy").
		Preload("ApprovedBy").
		First(&withdrawal, withdrawalID).Error; err != nil {
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
	withdrawal.ApprovedByID = &uid
	withdrawal.ApprovedBy = member
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

// DeleteMaterial deletes a specific material by ID.
func (mc *WithdrawalController) DeleteWithdraw(c *gin.Context) {
	materialID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid material ID"})
		return
	}

	if err := mc.DB.Delete(&models.Withdrawal{}, materialID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete withdrawal"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal deleted successfully"})
}
