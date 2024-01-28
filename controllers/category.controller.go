package controllers

import (
	"daijai/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CategoryController struct {
	DB *gorm.DB
}

// NewCategoryController creates a new instance of CategoryController.
func NewCategoryController(db *gorm.DB) *CategoryController {
	return &CategoryController{
		DB: db,
	}
}

// CreateCategory handles the creation of a new category.
func (mc *CategoryController) CreateCategory(c *gin.Context) {
	var category models.Category

	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := mc.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, category)
}

func (mc *CategoryController) GetCategories(c *gin.Context) {
	var categories []models.Category

	isFg := c.Query(models.MaterialType_Param) == models.MaterialType_FinishedGood

	if err := mc.DB.
		Preload("Materials").
		Where("is_fg = ?", isFg).
		Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetCategoryByID returns a specific category by ID.
func (mc *CategoryController) GetCategoryByID(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	var category models.Category
	if err := mc.DB.First(&category, categoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// UpdateCategory updates a specific category by ID.
func (mc *CategoryController) UpdateCategory(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	var existingCategory models.Category
	if err := mc.DB.First(&existingCategory, categoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	var updatedCategory models.Category
	if err := c.ShouldBindJSON(&updatedCategory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update only the fields you want to allow being updated.
	existingCategory.Slug = updatedCategory.Slug
	existingCategory.Title = updatedCategory.Title
	existingCategory.Subtitle = updatedCategory.Subtitle

	if err := mc.DB.Save(&existingCategory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	c.JSON(http.StatusOK, existingCategory)
}

// DeleteCategory deletes a specific category by ID.
func (mc *CategoryController) DeleteCategory(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	if err := mc.DB.Delete(&models.Category{}, categoryID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
