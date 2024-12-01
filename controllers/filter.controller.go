package controllers

import (
	"daijai/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FilterController struct {
	DB *gorm.DB
	BaseController
}

func NewFilterController(db *gorm.DB) *FilterController {
	return &FilterController{DB: db}
}

func (ctrl *FilterController) GetCategories(c *gin.Context) {
	var categories []models.Category
	if err := ctrl.DB.Find(&categories).Error; err != nil {
		ctrl.LogErrorAndSendBadRequest(c, "Failed to get categories")
		return
	}
	c.JSON(http.StatusOK, categories)
}

func (ctrl *FilterController) GetMaterialsByCategoryID(c *gin.Context) {
	categoryID := c.Param("id")
	showPrice := c.Query("showPrice")
	showStock := c.Query("showStock")
	showSupplier := c.Query("showSupplier")

	var materials []models.Material
	query := ctrl.DB.Debug().Where("category_id = ?", categoryID)

	allowFields := []string{"id", "slug", "title", "subtitle", "image_path", "category_id", "is_fg"}
	if showPrice == "true" {
		allowFields = append(allowFields, "default_price")
	}
	if showSupplier == "true" {
		allowFields = append(allowFields, "supplier")
	}
	query = query.Select(allowFields)

	if showStock == "true" {
		query = query.Preload("Sums")
	}
	if err := query.
		Order("id asc").
		Find(&materials).
		Error; err != nil {
		ctrl.LogErrorAndSendBadRequest(c, "Failed to get materials for category")
		return
	}
	c.JSON(http.StatusOK, materials)
}
