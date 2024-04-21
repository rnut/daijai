package controllers

import (
	"daijai/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

// DrawingController handles CRUD operations for the Drawing model.
type OrderController struct {
	DB *gorm.DB
	BaseController
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
	if err := odc.DB.Preload("BOMs.Material").First(&drawing, request.DrawingID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Drawing"})
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

	if err := odc.DB.Transaction(func(tx *gorm.DB) error {
		order.CreatedByID = member.ID
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		mainInventory := uint(1)

		for _, bom := range drawing.BOMs {
			// get InventoryMaterial that availableqty > 0 and not out of stock order by date created asc limit by sum of available == reserve
			var inventoryMaterials []models.InventoryMaterial
			if err := tx.
				Where("material_id = ? AND inventory_id = ? AND availabel_qty > 0 AND is_out_of_stock = false", bom.MaterialID, mainInventory).
				Order("created_at asc").
				Find(&inventoryMaterials).Error; err != nil {
				return err
			}

			target := bom.Quantity * order.ProducedQuantity

			var orderBom models.OrderBom
			orderBom.OrderID = order.ID
			orderBom.BOMID = bom.ID
			orderBom.TargetQty = target
			orderBom.IsCompletelyWithdraw = false
			orderBom.WithdrawedQty = 0
			if err := tx.Create(&orderBom).Error; err != nil {
				return err
			}

			totalReserve := int64(0)
			var isFullFilled bool
			for _, mat := range inventoryMaterials {
				if isFullFilled {
					break
				}
				// calculate target and available qty
				requiredQty := target - totalReserve
				availabelQty := mat.AvailableQty
				// calculate reserve qty
				var rQty int64
				var isInventoryOutOfStock bool
				if availabelQty <= requiredQty {
					rQty = availabelQty
					isInventoryOutOfStock = true
				} else {
					rQty = requiredQty
					isInventoryOutOfStock = false
				}
				totalReserve += rQty
				if totalReserve == target {
					isFullFilled = true
				}

				var updatedReserve = mat.Reserve + rQty
				// create inventory material transaction
				var transaction models.InventoryMaterialTransaction
				transaction.InventoryMaterialID = mat.ID
				transaction.Quantity = rQty
				transaction.InventoryType = models.InventoryType_RESERVE
				transaction.InventoryTypeDescription = models.InventoryTypeDescription_ORDER
				transaction.ExistingQuantity = mat.Quantity
				transaction.ExistingReserve = mat.Reserve
				transaction.UpdatedQuantity = mat.Quantity
				transaction.UpdatedReserve = updatedReserve
				transaction.OrderID = &order.ID
				if err := tx.Create(&transaction).Error; err != nil {
					return err
				}

				// create order reserving
				var orderReserving models.OrderReserving
				orderReserving.OrderID = order.ID
				orderReserving.OrderBomID = orderBom.ID
				orderReserving.ReceiptID = *mat.ReceiptID
				orderReserving.InventoryMaterialID = mat.ID
				orderReserving.Quantity = rQty
				orderReserving.Status = models.OrderReservingStatus_Reserved

				if err := tx.Create(&orderReserving).Error; err != nil {
					return err
				}

				// update inventory material and out of stock
				mat.Reserve = updatedReserve
				mat.AvailableQty -= rQty
				mat.IsOutOfStock = isInventoryOutOfStock
				if err := tx.Save(&mat).Error; err != nil {
					return err
				}

				matID := mat.MaterialID
				invID := mat.InventoryID
				odc.SumMaterial(tx, "order", matID, invID)
			}

			orderBom.ReservedQty = totalReserve
			orderBom.IsFullFilled = isFullFilled
			if err := tx.Save(&orderBom).Error; err != nil {
				return err
			}

			if !isFullFilled {
				var sg models.PurchaseSuggestion
				sg.OrderBomID = orderBom.ID
				sg.Status = models.PurchaseSuggestionStatus_Ready
				if err := tx.Create(&sg).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Order"})
		log.Println(err)
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
		Preload("OrderBoms").
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
		Preload("OrderBOMs.BOM.Material").
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

func (odc *OrderController) GetOrderBOMBySlug(c *gin.Context) {
	slug := c.Param("slug")

	// Get the order
	var order models.Order
	if err := odc.
		DB.
		Preload("OrderBOMs.BOM.Material").
		Where("slug = ?", slug).
		First(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Order"})
		return
	}
	c.JSON(http.StatusOK, order.OrderBoms)

}
