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

		// Update the associated materials' quantity
		for i, rm := range request.ReceiptMaterials {
			material := models.Material{}
			if err := rc.DB.First(&material, rm.MaterialID).Error; err != nil {
				return err
			}

			// Update the material's quantity
			if material.IncomingQuantity >= rm.Quantity {
				material.IncomingQuantity -= rm.Quantity
				material.Quantity += rm.Quantity
			} else {
				q := rm.Quantity - material.IncomingQuantity
				material.IncomingQuantity = 0
				material.Quantity += q
			}

			if err := tx.Save(&material).Error; err != nil {
				tx.Rollback()
				return err
			}

			// set material back to withdrawalMaterials
			request.ReceiptMaterials[i].Material = material
		}
		return nil
	}); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create drawing"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Receipt created successfully", "receipt": request})
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
		Preload("ReceiptMaterials").
		Preload("CreatedBy")

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

	// Update Receipt fields here if needed
	// ...

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
