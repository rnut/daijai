package controllers

import (
	"daijai/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MaterialController struct {
	DB *gorm.DB
	BaseController
}

// NewMaterialController creates a new instance of MaterialController.
func NewMaterialController(db *gorm.DB) *MaterialController {
	return &MaterialController{
		DB: db,
	}
}

func (mc *MaterialController) CreateMaterial(c *gin.Context) {
	var material models.Material
	if err := c.ShouldBindJSON(&material); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a new material
	if err := mc.DB.Create(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := mc.DB.Preload("Category").First(&material, material.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Material Category"})
		return
	}

	c.JSON(http.StatusCreated, material)
}

// GetMaterials returns a list of all materials.
func (mc *MaterialController) GetMaterials(c *gin.Context) {
	var categories []models.Category
	isFg := c.Query(models.MaterialType_Param) == models.MaterialType_FinishedGood
	if err := mc.DB.
		Preload("Materials", func(db *gorm.DB) *gorm.DB {
			return db.Order("materials.id ASC")
		}).
		Preload("Materials.Sum").
		// mainInventoryID := uint(1)
		// Preload("Materials.Sum", "inventory_id = ?", mainInventoryID). // sum only main inventory
		Where("is_fg = ?", isFg).
		Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve categories"})
		return
	}

	// get inventories
	var inventories []models.Inventory
	if err := mc.DB.Find(&inventories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve inventories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories, "inventories": inventories})
}

// get deleted materials
func (mc *MaterialController) GetDeletedMaterials(c *gin.Context) {
	var materials []models.Material
	if err := mc.DB.Unscoped().Where("deleted_at IS NOT NULL").Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve materials"})
		return
	}

	c.JSON(http.StatusOK, materials)
}

// restore deleted material
func (mc *MaterialController) RestoreMaterial(c *gin.Context) {
	materialID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid material ID"})
		return
	}

	if err := mc.DB.Unscoped().Model(&models.Material{}).Where("id = ?", materialID).Update("deleted_at", nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore material"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Material restored successfully"})
}

// permanently delete material
func (mc *MaterialController) PermanentlyDeleteMaterial(c *gin.Context) {
	materialID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid material ID"})
		return
	}

	fmt.Println("materialID", materialID)

	if err := mc.DB.Unscoped().Delete(&models.Material{}, materialID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete material"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Material deleted permanently"})
}

// GetMaterialByID returns a specific material by ID.
func (mc *MaterialController) GetMaterialBySlug(c *gin.Context) {
	slug := c.Param("slug")
	var material models.Material
	if err := mc.
		DB.
		Preload("Category").
		Where("slug = ?", slug).
		First(&material).
		Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material not found"})
		return
	}

	// get inventories and transactions
	var inventories []models.Inventory
	if err := mc.
		DB.
		Preload("Transactions", "material_id = ?", material.ID, func(db *gorm.DB) *gorm.DB {
			return db.Order("transactions.id DESC")
		}).
		Preload("Transactions.Receipt").
		Find(&inventories).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve inventories"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"material":    material,
		"inventories": inventories,
	})
}

// UpdateMaterial updates a specific material by ID.
func (mc *MaterialController) UpdateMaterial(c *gin.Context) {
	materialID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid material ID"})
		return
	}

	var material models.Material
	if err := c.ShouldBindJSON(&material); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingMaterial models.Material
	if err := mc.DB.Preload("Sums").First(&existingMaterial, materialID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material not found"})
		return
	}

	existingMaterial.CategoryID = material.CategoryID
	existingMaterial.Slug = material.Slug
	existingMaterial.Title = material.Title
	existingMaterial.Subtitle = material.Subtitle
	existingMaterial.ImagePath = material.ImagePath

	if err := mc.DB.Save(&existingMaterial).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update material"})
		return
	}

	c.JSON(http.StatusOK, existingMaterial)
}

// DeleteMaterial deletes a specific material by ID.
func (mc *MaterialController) DeleteMaterial(c *gin.Context) {
	materialID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid material ID"})
		return
	}

	if err := mc.DB.Delete(&models.Material{}, materialID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete material"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Material deleted successfully"})
}

// AdjustMaterialQuantity adjusts the quantity of a specific material by ID.
func (mc *MaterialController) AdjustMaterialQuantity(c *gin.Context) {
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

	materialID := c.Param("id")
	var material models.Material
	if err := mc.DB.First(&material, materialID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material not found"})
		return
	}

	// Get the adjustment value from the request body
	var req struct {
		Quantity     int64
		InventoryID  uint
		PricePerUnit int64
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid adjustment value"})
		return
	}

	if err := mc.DB.Transaction(func(tx *gorm.DB) error {
		// Create a new adjustment record
		adjustment := models.Adjustment{
			Quantity:     req.Quantity,
			InventoryID:  req.InventoryID,
			MaterialID:   material.ID,
			CreatedByID:  member.ID,
			PricePerUnit: req.PricePerUnit,
		}
		if err := tx.Create(&adjustment).Error; err != nil {
			return err
		}

		// create inventory material
		inventoryMaterial := models.InventoryMaterial{
			InventoryID:  req.InventoryID,
			MaterialID:   material.ID,
			AdjustmentID: &adjustment.ID,
			Quantity:     req.Quantity,
			AvailabelQty: req.Quantity,
			IsOutOfStock: false,
			Price:        adjustment.PricePerUnit,
		}
		if err := tx.Create(&inventoryMaterial).Error; err != nil {
			return err
		}

		// count
		var counter struct {
			Quantity   int64
			Reserved   int64
			Withdrawed int64
		}
		if err := tx.
			Model(&models.InventoryMaterial{}).
			Select("SUM(quantity) as quantity, SUM(reserve) as reserved, SUM(withdrawed) as withdrawed").
			Where("material_id = ?", material.ID).
			Where("is_out_of_stock = ?", false).
			Find(&counter).Error; err != nil {
			return err
		}

		// update sum material inventory
		mc.SumMaterial(tx, "receipt", material.ID, req.InventoryID)

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// reload material
	if err := mc.DB.Preload("Sums").First(&material, materialID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Material Category"})
		return
	}

	c.JSON(http.StatusOK, material)
}
