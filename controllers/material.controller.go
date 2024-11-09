package controllers

import (
	"daijai/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

// create search material by title or subtitle or supplier or slug
func (mc *MaterialController) SearchMaterials(c *gin.Context) {
	search := strings.ToLower(c.Query("q"))
	isFG := c.Query(models.MaterialType_Param) == models.MaterialType_FinishedGood
	var materials []models.Material
	count := int64(0)

	query := mc.
		DB.
		Where("LOWER(title) LIKE LOWER(?) OR LOWER(subtitle) LIKE LOWER(?) OR LOWER(supplier) LIKE LOWER(?) OR LOWER(slug) LIKE LOWER(?)", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%").
		Where("is_fg = ?", isFG)

	if err := query.
		Model(&models.Material{}).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count materials"})
		return
	}

	if err := query.
		Preload("Category").
		Where("is_fg = ?", isFG).
		Order("materials.id ASC").
		Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve materials"})
		return
	}

	category := models.Category{
		Slug:      "",
		Title:     fmt.Sprintf("Search result `%s`", search),
		Subtitle:  "Number of Results: " + strconv.FormatInt(count, 10),
		Materials: materials,
	}

	result := struct {
		Query           string
		NumberOfResults int64
		Categories      []models.Category
	}{
		Query:           search,
		NumberOfResults: count,
		Categories:      []models.Category{category},
	}

	c.JSON(http.StatusOK, result)
}

// Query materials by category, inventory, and material type
func (mc *MaterialController) QueryMaterials(c *gin.Context) {
	categoryID := c.Query("categoryID")
	inventoryIDs := c.Query("inventoryIDs")
	isFg := c.Query(models.MaterialType_Param) == models.MaterialType_FinishedGood
	distinct := c.Query("distinct") == "true"

	// split inventory slugs by delimeter ','
	var inventorySlugArr []string
	if inventoryIDs != "" {
		inventorySlugArr = strings.Split(inventoryIDs, ",")
	}

	var materials []models.Material
	if err := mc.DB.
		Preload("Sums", "inventory_id IN ?", inventorySlugArr).
		Where("category_id = ?", categoryID).
		Where("is_fg = ?", isFg).
		Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve materials"})
		return
	}

	if distinct {
		var filteredMaterials []models.Material
		for _, material := range materials {
			if len(*material.Sums) > 0 {
				filteredMaterials = append(filteredMaterials, material)
			}
		}
		c.JSON(http.StatusOK, filteredMaterials)
	} else {
		c.JSON(http.StatusOK, materials)
	}

}

// GetMaterials returns a list of all materials.
func (mc *MaterialController) GetMaterials(c *gin.Context) {
	var categories []models.Category
	isFg := c.Query(models.MaterialType_Param) == models.MaterialType_FinishedGood
	fmt.Println("isFg", isFg)
	if err := mc.DB.
		Preload("Materials", func(db *gorm.DB) *gorm.DB {
			return db.Order("materials.id ASC")
		}).
		Preload("Materials.Sums").
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
	existingMaterial.Supplier = material.Supplier
	existingMaterial.DefaultPrice = material.DefaultPrice
	existingMaterial.Max = material.Max
	existingMaterial.Min = material.Min

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
			InventoryID:           req.InventoryID,
			MaterialID:            material.ID,
			AdjustmentID:          &adjustment.ID,
			Quantity:              req.Quantity,
			AvailableQty:          req.Quantity,
			IsOutOfStock:          false,
			Price:                 adjustment.PricePerUnit,
			InventoryMaterialType: models.InventoryMaterialType_Adjust,
		}
		if err := tx.Create(&inventoryMaterial).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// update sum material inventory
	mc.SumMaterial(mc.DB, "adjust", material.ID, req.InventoryID)

	// reload material
	if err := mc.DB.Preload("Sums").First(&material, materialID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Material Category"})
		return
	}

	c.JSON(http.StatusOK, material)
}
