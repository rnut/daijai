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

// CreateMaterial handles the creation of a new material.
func (mc *MaterialController) CreateMaterial(c *gin.Context) {
	var material models.Material

	if err := c.ShouldBindJSON(&material); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := mc.DB.Create(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create material"})
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
