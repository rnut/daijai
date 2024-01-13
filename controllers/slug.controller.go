package controllers

import (
	"daijai/models"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SlugController struct {
	DB *gorm.DB
	BaseController
}

func NewSlugController(db *gorm.DB) *SlugController {
	return &SlugController{
		DB: db,
	}
}

func (slc *SlugController) GetSlug(c *gin.Context) {
	var slug models.Slugger
	id := c.Param("slug")
	if err := slc.
		DB.
		Where("table_name = ?", id).
		First(&slug).
		Error; err != nil {
		c.JSON(404, gin.H{"error": "Slug not found"})
		return
	}
	c.JSON(200, slug)
}

func (slc *SlugController) GetAllSluggers(c *gin.Context) {
	var sluggers []models.Slugger
	if err := slc.DB.Find(&sluggers).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to get sluggers"})
		return
	}
	c.JSON(200, sluggers)
}

func (slc *SlugController) RequestSlug(c *gin.Context) {
	var slug models.Slugger
	id := c.Param("slug")
	if err := slc.
		DB.
		Where("table_name = ?", id).
		First(&slug).
		Error; err != nil {
		c.JSON(404, gin.H{"error": "Slug not found"})
		return
	}

	// Update the slug value
	slug.Value++
	if err := slc.DB.Save(&slug).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update slug value"})
		return
	}

	// Combine prefix and updated value

	pad := "%0" + strconv.Itoa(slug.Pad) + "d"
	incrementer := fmt.Sprintf(pad, slug.Value)
	combinedValue := slug.Prefix + incrementer

	c.JSON(200, gin.H{"slug": combinedValue})
}
