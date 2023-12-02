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
	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	drw.CreatedByID = member.ID
	drw.CreatedBy = member
	drw.Slug = c.Request.FormValue("Slug")
	drw.PartNumber = c.Request.FormValue("PartNumber")
	pQty, err := strconv.ParseInt(c.Request.FormValue("ProducedQuantity"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ProducedQuantity"})
		return
	}
	drw.ProducedQuantity = pQty

	if err := dc.DB.Transaction(func(tx *gorm.DB) error {
		// Save uploaded image
		_, header, err := c.Request.FormFile("image")
		if err != nil {
			return err
		}
		path := "/drawings/" + drw.Slug + ".jpg"
		filePath := "./public" + path
		if err := c.SaveUploadedFile(header, filePath); err != nil {
			return err
		}

		drw.ImagePath = path
		if err := tx.Create(&drw).Error; err != nil {
			return err
		}
		var bombs []models.Bomb
		mIDs := c.PostFormArray("Bombs.MaterialID")
		qts := c.PostFormArray("Bombs.Quantity")
		prices := c.PostFormArray("Bombs.Price")

		for i := 0; i < len(mIDs); i++ {
			mID, err := strconv.ParseUint(mIDs[i], 10, 64)
			if err != nil {
				break
			}
			qty, err := strconv.ParseInt(qts[i], 10, 64)
			if err != nil {
				break
			}
			price, err := strconv.ParseInt(prices[i], 10, 64)
			if err != nil {
				break
			}
			b := models.Bomb{
				DrawingID:  drw.ID,
				Quantity:   qty,
				MaterialID: uint(mID),
				Price:      price,
			}
			if err := tx.Save(&b).Error; err != nil {
				return err
			}
			bombs = append(bombs, b)
		}
		drw.Bombs = bombs

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create drawing"})
		return
	}
	c.JSON(http.StatusCreated, drw)
}

// GetDrawings returns a list of all drawings.
func (dc *DrawingController) GetDrawings(c *gin.Context) {
	var drawings []models.Drawing

	if err := dc.DB.Preload("Bombs").Preload("Bombs.Material").Preload("CreatedBy").Find(&drawings).Error; err != nil {
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
	if err := dc.DB.Preload("Bombs.Material.Category").Preload("CreatedBy").First(&drawing, drawingID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Drawing not found"})
		return
	}

	c.JSON(http.StatusOK, drawing)
}

// / NOT ALLOW FOR UPDATE
// UpdateDrawing updates a specific drawing by ID.
// func (dc *DrawingController) UpdateDrawing(c *gin.Context) {
// 	drawingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid drawing ID"})
// 		return
// 	}

// 	var existingDrawing models.Drawing
// 	if err := dc.DB.First(&existingDrawing, drawingID).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Drawing not found"})
// 		return
// 	}

// 	var updatedDrawing models.Drawing
// 	if err := c.ShouldBindJSON(&updatedDrawing); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Update only the fields you want to allow being updated.
// 	existingDrawing.Slug = updatedDrawing.Slug
// 	existingDrawing.PartNumber = updatedDrawing.PartNumber
// 	existingDrawing.ProducedQuantity = updatedDrawing.ProducedQuantity

// 	if err := dc.DB.Save(&existingDrawing).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update drawing"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, existingDrawing)
// }

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
