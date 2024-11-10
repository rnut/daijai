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
	var materials []models.Material
	if err := ctrl.DB.Where("category_id = ?", categoryID).Find(&materials).Error; err != nil {
		ctrl.LogErrorAndSendBadRequest(c, "Failed to get materials for category")
		return
	}
	c.JSON(http.StatusOK, materials)
}
