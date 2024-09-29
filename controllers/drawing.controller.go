package controllers

import (
	"daijai/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

// DrawingController handles CRUD operations for the Drawing model.
type DrawingController struct {
	DB *gorm.DB
	BaseController
}

// NewDrawingController creates a new instance of DrawingController.
func NewDrawingController(db *gorm.DB) *DrawingController {
	return &DrawingController{
		DB: db,
	}
}

func (dc *DrawingController) GetNewDrawingInfo(c *gin.Context) {
	isFg := c.Param("type") == models.MaterialType_FinishedGood
	var response struct {
		Slug       string
		Categories []models.Category
	}

	var categories []models.Category
	if err := dc.DB.
		Preload("Materials.Sums").
		Where("is_fg = ?", isFg).
		Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Categories"})
		return
	}

	// get slug
	if err := dc.RequestSlug(&response.Slug, dc.DB, "drawings"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Slug"})
		return
	}

	response.Categories = categories
	c.JSON(http.StatusOK, response)
}

func (dc *DrawingController) CreateDrawing(c *gin.Context) {
	var uid uint
	if err := dc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := dc.getUserDataByUserID(dc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var drw models.Drawing
	if err := c.ShouldBindJSON(&drw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// create drawing
	drw.CreatedByID = member.ID
	drw.CreatedBy = member
	if err := dc.DB.Create(&drw).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create drawing"})
		return
	}

	c.JSON(http.StatusCreated, drw)
}

// GetDrawings returns a list of all drawings.
func (dc *DrawingController) GetDrawings(c *gin.Context) {
	var drawings []models.Drawing

	drawingType := c.Query(models.MaterialType_Param)
	isFG := drawingType == models.MaterialType_FinishedGood

	// Use the drawingType value as needed

	if err := dc.
		DB.
		Preload("BOMs.Material.Category").
		Preload("CreatedBy").
		Where("is_fg = ?", isFG).
		Find(&drawings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve drawings"})
		return
	}

	c.JSON(http.StatusOK, drawings)
}

// GetDrawingByID returns a specific drawing by ID.
func (dc *DrawingController) GetDrawingByID(c *gin.Context) {
	drawingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid drawing ID"})
		return
	}

	var drawing models.Drawing
	if err := dc.DB.
		Preload("BOMs.Material.Category").
		Preload("BOMs.Material.Sums").
		Preload("CreatedBy").
		First(&drawing, drawingID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Drawing not found"})
		return
	}

	c.JSON(http.StatusOK, drawing)
}

func (dc *DrawingController) UpdateDrawing(c *gin.Context) {
	drawingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid drawing ID"})
		return
	}
	var drw models.Drawing
	if err := dc.DB.Preload("BOMs").First(&drw, drawingID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Drawing not found"})
		return
	}

	// UPDATE Drawing fields
	var req models.Drawing
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var uid uint
	if err := dc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := dc.getUserDataByUserID(dc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// UPDATE Drawing fields
	drw.ImagePath = req.ImagePath
	drw.Slug = req.Slug
	drw.PartNumber = req.PartNumber
	drw.IsFG = req.IsFG

	if err := dc.DB.Transaction(func(tx *gorm.DB) error {
		// DELETE EXISTING BOMs
		if err := tx.Where("drawing_id = ?", drw.ID).Delete(&models.BOM{}).Error; err != nil {
			return err
		}

		// CREATE NEW BOMs
		for _, v := range req.BOMs {
			v.DrawingID = drw.ID
			if err := tx.Save(&v).Error; err != nil {
				return err
			}
		}
		drw.BOMs = req.BOMs
		if err := tx.Save(&drw).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create drawing"})
		return
	}
	c.JSON(http.StatusCreated, drw)
}

// DeleteDrawing deletes a specific drawing by ID.
func (dc *DrawingController) DeleteDrawing(c *gin.Context) {
	drawingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid drawing ID"})
		return
	}

	if err := dc.DB.Delete(&models.Drawing{}, drawingID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete drawing"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Drawing deleted successfully"})
}
