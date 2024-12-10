package controllers

import (
	"daijai/models"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

type ExtendOrdererController struct {
	DB *gorm.DB
	BaseController
}

func NewExtendOrdererController(db *gorm.DB) *ExtendOrdererController {
	return &ExtendOrdererController{
		DB: db,
	}
}

func (wc *ExtendOrdererController) GetExtendOrders(c *gin.Context) {
	var orders []models.ExtendOrder
	if err := wc.
		DB.
		Preload("Order").
		Preload("Project").
		Preload("ExtendOrderBOMs.Material").
		Preload("CreatedBy").
		Find(&orders).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ExtendOrders"})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (wc *ExtendOrdererController) GetNewInfo(c *gin.Context) {
	var projects []models.Project
	if err := wc.DB.Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Projects"})
		return
	}

	var categories []models.Category
	if err := wc.DB.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Categories"})
		return
	}

	var materials []models.Material
	if err := wc.
		DB.
		Preload("Sums").
		Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Materials"})
		return
	}

	// get slug
	var slug string
	if err := wc.RequestSlug(&slug, wc.DB, "extend_orders"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Slug", "detail": err.Error()})
		return
	}
	var result struct {
		Slug       string
		Projects   []models.Project
		Categories []models.Category
		Materials  []models.Material
	}
	result.Materials = materials
	result.Categories = categories
	result.Projects = projects
	result.Slug = slug

	c.JSON(http.StatusOK, result)
}

func (odc *ExtendOrdererController) GetExtendOrderBySlug(c *gin.Context) {
	var order models.ExtendOrder
	slug := c.Param("slug")
	if err := odc.
		DB.
		Preload("ExtendOrderBOMs.Material").
		Preload("Project").
		Preload("Order").
		Preload("CreatedBy").
		Where("slug = ?", slug).
		First(&order).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ExtendOrder"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (odc *ExtendOrdererController) CreateExtendOrders(c *gin.Context) {
	var uid uint
	if err := odc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := odc.getUserDataByUserID(odc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.ExtendOrder
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order.CreatedByID = member.ID

	if err := odc.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		errorMsg := fmt.Sprintf("Failed to create ExtendOrders: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ExtendOrders"})
		log.Println(errorMsg)
		return
	}
	c.JSON(http.StatusCreated, order)
}
