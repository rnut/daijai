package controllers

import (
	"daijai/models"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

// DrawingController handles CRUD operations for the Drawing model.
type OrderController struct {
	DB *gorm.DB
	BaseController
	DebugController
}

// NewDrawingController creates a new instance of DrawingController.
func NewOrderController(db *gorm.DB) *OrderController {
	return &OrderController{
		DB: db,
	}
}

func (odc *OrderController) CreateOrder(c *gin.Context) {
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

	var request struct {
		Slug             string
		DrawingID        int64
		ProducedQuantity int64
		ProjectID        int64
		Notes            string
		IsFG             bool
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var drawing models.Drawing
	if err := odc.
		DB.
		Preload("BOMs.Material").
		First(&drawing, request.DrawingID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Drawing"})
		return
	}

	odc.PrintJSON(drawing)

	// check if drawig material is null
	var nullMaterials []uint
	for _, bom := range drawing.BOMs {
		if bom.Material == nil {
			nullMaterials = append(nullMaterials, bom.MaterialID)
		}
	}
	if len(nullMaterials) > 0 {
		var mats []models.Material
		if err := odc.DB.Unscoped().Find(&mats, nullMaterials).Error; err != nil {
			log.Printf("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Drawing has null material", "material_id": nullMaterials})
			return
		}
		errorMessage := "Materials-IDs: "
		for _, v := range mats {
			errorMessage += fmt.Sprintf("[%s] ", v.Slug)
		}

		errorMessage += "has some problem. Please check it out at materials page"
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": errorMessage})
		return
	}

	var order models.Order
	order.Slug = request.Slug
	order.DrawingID = drawing.ID
	order.ProducedQuantity = request.ProducedQuantity
	order.Drawing = drawing
	order.ProjectID = uint(request.ProjectID)
	order.Notes = request.Notes
	order.IsFG = request.IsFG
	order.CreatedByID = member.ID

	if err := odc.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		for _, bom := range drawing.BOMs {

			target := bom.Quantity * order.ProducedQuantity

			var orderBom models.OrderBOM
			orderBom.OrderID = order.ID
			orderBom.MaterialID = bom.MaterialID
			orderBom.DrawingID = bom.DrawingID
			orderBom.TargetQty = target
			orderBom.IsCompletelyWithdraw = false
			orderBom.WithdrawedQty = 0
			if err := tx.Create(&orderBom).Error; err != nil {
				return err
			}

			// get material
			var material models.Material
			if err := tx.Preload("Sums").First(&material, bom.MaterialID).Error; err != nil {
				return err
			}
			var materialAvialableQty int64 = 0
			for _, v := range *material.Sums {
				materialAvialableQty += v.Quantity
			}
			if materialAvialableQty < target {
				var sg models.PurchaseSuggestion
				sg.OrderBOMID = orderBom.ID
				sg.Status = models.PurchaseSuggestionStatus_Ready
				if err := tx.Create(&sg).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		log.Printf("Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create Order"})
		return
	}
	c.JSON(http.StatusCreated, order)
}

// / get all orders
func (odc *OrderController) GetOrders(c *gin.Context) {
	materialType := c.Query(models.MaterialType_Param)
	isFG := materialType == models.MaterialType_FinishedGood

	var orders []models.Order
	if err := odc.
		DB.
		Preload("Drawing").
		Preload("CreatedBy").
		Preload("OrderBOMs").
		Where("is_fg = ?", isFG).
		Find(&orders).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Orders"})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (odc *OrderController) GetOrderBySlug(c *gin.Context) {
	var order models.Order
	slug := c.Param("slug")
	if err := odc.
		DB.
		Preload("OrderBOMs.Material").
		Preload("Drawing").
		Preload("Project").
		Preload("CreatedBy").
		Where("slug = ?", slug).
		First(&order).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Order"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (odc *OrderController) GetNewOrderInfo(c *gin.Context) {
	materialType := c.Query(models.MaterialType_Param)
	isFG := materialType == models.MaterialType_FinishedGood

	// get projects
	var projects []models.Project
	if err := odc.
		DB.
		Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Projects"})
		return
	}
	// get drawings
	var drawings []models.Drawing
	if err := odc.
		DB.
		Where("is_fg = ?", isFG).
		Preload("BOMs.Material.Sums").
		Find(&drawings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Drawings"})
		return
	}

	// get slug
	var slug string
	if err := odc.RequestSlug(&slug, odc.DB, "orders"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Slug", "detail": err.Error()})
		return
	}

	var response struct {
		Slug     string
		Projects []models.Project
		Drawings []models.Drawing
	}
	response.Slug = slug
	response.Projects = projects
	response.Drawings = drawings
	c.JSON(http.StatusOK, response)
}

func (odc *OrderController) GetOrderInfo(c *gin.Context) {
	slug := c.Param("slug")

	// Get the order
	var order models.Order
	if err := odc.
		DB.
		Preload("OrderBOMs.Material").
		Preload("OrderReservings.InventoryMaterial.Material").
		Where("slug = ?", slug).
		First(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Order"})
		return
	}
	c.JSON(http.StatusOK, order)

}
