package controllers

import (
	"daijai/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MaterialController struct {
	DB *gorm.DB
}

// NewMaterialController creates a new instance of MaterialController.
func NewMaterialController(db *gorm.DB) *MaterialController {
	return &MaterialController{
		DB: db,
	}
}

func (mc *MaterialController) CreateMaterial(c *gin.Context) {
	var material models.Material
	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get form values
	cID, err := strconv.ParseUint(c.Request.FormValue("CategoryID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CategoryID"})
		return
	}
	material.CategoryID = uint(cID)

	material.Slug = c.Request.FormValue("Slug")
	material.Title = c.Request.FormValue("Title")

	material.Subtitle = c.Request.FormValue("Subtitle")

	min, err := strconv.ParseInt(c.Request.FormValue("Min"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Min"})
		return
	}
	material.Min = min

	max, err := strconv.ParseInt(c.Request.FormValue("Max"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Max"})
		return
	}
	material.Max = max

	isFG, _ := strconv.ParseBool(c.Request.FormValue("IsFG"))
	material.IsFG = isFG

	// check form has image value
	if _, header, err := c.Request.FormFile("image"); err == nil {
		// Save uploaded image
		path := "/materials/" + material.Slug + ".jpg"
		filePath := "./public" + path
		if err := c.SaveUploadedFile(header, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			return
		}
		material.ImagePath = path
	}
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
	mainInventoryID := uint(1)
	isFg := c.Query(models.MaterialType_Param) == models.MaterialType_FinishedGood
	if err := mc.DB.
		Preload("Materials", func(db *gorm.DB) *gorm.DB {
			return db.Order("materials.id ASC")
		}).
		Preload("Materials.Sum", "inventory_id = ?", mainInventoryID).
		Where("is_fg = ?", isFg).
		Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
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
	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	materialID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid material ID"})
		return
	}

	var existingMaterial models.Material
	if err := mc.DB.First(&existingMaterial, materialID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material not found"})
		return
	}

	// Get form values
	cID, err := strconv.ParseUint(c.Request.FormValue("CategoryID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CategoryID"})
		return
	}
	existingMaterial.CategoryID = uint(cID)
	existingMaterial.Slug = c.Request.FormValue("Slug")
	existingMaterial.Title = c.Request.FormValue("Title")
	existingMaterial.Subtitle = c.Request.FormValue("Subtitle")

	// min, err := strconv.ParseInt(c.Request.FormValue("Min"), 10, 64)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Min"})
	// 	return
	// }
	// existingMaterial.Min = min
	// max, err := strconv.ParseInt(c.Request.FormValue("Max"), 10, 64)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Max"})
	// 	return
	// }
	// existingMaterial.Max = max
	// existingMaterial.Price = c.Request.FormValue("Price")
	if _, header, err := c.Request.FormFile("image"); err == nil {
		// Save uploaded image
		path := "/materials/" + existingMaterial.Slug + ".jpg"
		filePath := "./public" + path
		if err := c.SaveUploadedFile(header, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			return
		}

		existingMaterial.ImagePath = path
	}
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
