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
		DrawingID        int64
		ProducedQuantity int64
		ProjectID        int64
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var drawing models.Drawing
	if err := odc.DB.Preload("Boms.Material").First(&drawing, request.DrawingID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Drawing"})
		return
	}

	var order models.Order
	order.DrawingID = drawing.ID
	order.ProducedQuantity = request.ProducedQuantity
	order.Drawing = drawing
	order.ProjectID = uint(request.ProjectID)

	if err := odc.DB.Transaction(func(tx *gorm.DB) error {
		order.CreatedByID = member.ID
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		for _, bom := range drawing.Boms {
			mainInventory := uint(1)
			var transaction models.Transaction
			if err := tx.
				Where("material_id = ? AND inventory_id = ?", bom.MaterialID, mainInventory).
				Last(&transaction).
				Error; err != nil {
				return err
			}
			availabelQty := transaction.TotalQuantity - transaction.TotalReserve
			if availabelQty < 0 {
				availabelQty = 0
			}

			var reserve int64
			requireQty := bom.Quantity * order.ProducedQuantity
			if availabelQty < requireQty {
				reserve = availabelQty
			} else {
				reserve = requireQty
			}

			log.Println("reserve", reserve)

			t := models.Transaction{
				MaterialID:     bom.MaterialID,
				InventoryID:    mainInventory,
				Quantity:       transaction.TotalQuantity,
				Reserve:        transaction.TotalReserve,
				QuantityChange: 0,
				ReserveChange:  reserve,
				TotalQuantity:  transaction.TotalQuantity,
				TotalReserve:   transaction.TotalReserve + reserve,
				Type:           "reserve",
			}
			if err := tx.Create(&t).Error; err != nil {
				return err
			}

			// get InventoryMaterial that availableqty > 0 and not out of stock order by date created asc limit by sum of available == reserve
			var inventoryMaterials []models.InventoryMaterial
			if err := tx.
				Where("material_id = ? AND availabel_qty > 0 AND is_out_of_stock = false", bom.MaterialID).
				Order("created_at asc").
				Find(&inventoryMaterials).Error; err != nil {
				return err
			}

			totalReserve := int64(0)
			for _, mat := range inventoryMaterials {
				requiredQty := reserve - totalReserve
				availabelQty := mat.AvailabelQty

				var qty int64
				var isFullFilled bool
				if availabelQty < requiredQty {
					qty = availabelQty
					isFullFilled = false
				} else {
					qty = requiredQty
					isFullFilled = true
				}
				totalReserve += qty

				var orderBom models.OrderBom
				orderBom.OrderID = order.ID
				orderBom.BomID = bom.ID
				orderBom.Reserved = qty
				orderBom.IsFullFilled = isFullFilled
				orderBom.InventoryMaterialID = mat.ID

				if err := tx.Create(&orderBom).Error; err != nil {
					return err
				}

				// update inventory material
				mat.Reserve += qty
				mat.AvailabelQty -= qty
				if mat.Reserve == mat.AvailabelQty {
					mat.IsOutOfStock = true
				}
				if err := tx.Save(&mat).Error; err != nil {
					return err
				}
			}

		}
		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create drawing"})
		log.Println(err)
		return
	}
	c.JSON(http.StatusCreated, request)
}

// / get all orders
func (odc *OrderController) GetOrders(c *gin.Context) {
	var orders []models.Order
	if err := odc.
		DB.
		Find(&orders).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Orders"})
		return
	}
	c.JSON(http.StatusOK, orders)
}
