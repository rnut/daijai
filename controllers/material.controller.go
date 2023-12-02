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
	price, err := strconv.ParseInt(c.Request.FormValue("Price"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Price"})
		return
	}
	material.Price = price

	qty, err := strconv.ParseInt(c.Request.FormValue("Quantity"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Quantity"})
		return
	}
	material.Quantity = qty

	iuQt, err := strconv.ParseInt(c.Request.FormValue("InUseQuantity"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid InUseQuantity"})
		return
	}
	material.InUseQuantity = iuQt

	icQt, err := strconv.ParseInt(c.Request.FormValue("InUseQuantity"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid IncomingQuantity"})
		return
	}
	material.IncomingQuantity = icQt

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
	material.Supplier = c.Request.FormValue("Supplier")

	_, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image upload failed"})
		return
	}
	// Save uploaded image
	path := "/materials/" + material.Slug + ".jpg"
	filePath := "./public" + path
	if err := c.SaveUploadedFile(header, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	material.ImagePath = path

	if err := mc.DB.Create(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create material"})
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
	var materials []models.Material

	if err := mc.DB.Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve materials"})
		return
	}

	c.JSON(http.StatusOK, materials)
}

// GetMaterialByID returns a specific material by ID.
func (mc *MaterialController) GetMaterialByID(c *gin.Context) {
	materialID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid material ID"})
		return
	}

	var material models.Material
	if err := mc.DB.First(&material, materialID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material not found"})
		return
	}

	c.JSON(http.StatusOK, material)
}

// UpdateMaterial updates a specific material by ID.
func (mc *MaterialController) UpdateMaterial(c *gin.Context) {
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

	var updatedMaterial models.Material
	if err := c.ShouldBindJSON(&updatedMaterial); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update only the fields you want to allow being updated.
	existingMaterial.Slug = updatedMaterial.Slug
	existingMaterial.Title = updatedMaterial.Title
	existingMaterial.Subtitle = updatedMaterial.Subtitle
	existingMaterial.Price = updatedMaterial.Price
	existingMaterial.Quantity = updatedMaterial.Quantity
	existingMaterial.InUseQuantity = updatedMaterial.InUseQuantity
	existingMaterial.Supplier = updatedMaterial.Supplier
	existingMaterial.Min = updatedMaterial.Min
	existingMaterial.Max = updatedMaterial.Max

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
