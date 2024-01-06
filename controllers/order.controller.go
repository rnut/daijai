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
	order.Slug = request.Slug
	order.DrawingID = drawing.ID
	order.ProducedQuantity = request.ProducedQuantity
	order.Drawing = drawing
	order.ProjectID = uint(request.ProjectID)

	if err := odc.DB.Transaction(func(tx *gorm.DB) error {
		order.CreatedByID = member.ID
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		mainInventory := uint(1)

		for _, bom := range drawing.Boms {
			// var transaction models.AppLog
			// if err := tx.
			// 	Where("material_id = ? AND inventory_id = ?", bom.MaterialID, mainInventory).
			// 	Last(&transaction).
			// 	Error; err != nil {
			// 	return err
			// }
			// availabelQty := transaction.TotalQuantity - transaction.TotalReserve
			// if availabelQty < 0 {
			// 	availabelQty = 0
			// }

			// var reserve int64
			// requireQty := bom.Quantity * order.ProducedQuantity
			// if availabelQty < requireQty {
			// 	reserve = availabelQty
			// } else {
			// 	reserve = requireQty
			// }

			// t := models.AppLog{
			// 	MaterialID:     bom.MaterialID,
			// 	InventoryID:    mainInventory,
			// 	Quantity:       transaction.TotalQuantity,
			// 	Reserve:        transaction.TotalReserve,
			// 	QuantityChange: 0,
			// 	ReserveChange:  reserve,
			// 	TotalQuantity:  transaction.TotalQuantity,
			// 	TotalReserve:   transaction.TotalReserve + reserve,
			// 	Type:           "reserve",
			// }
			// if err := tx.Create(&t).Error; err != nil {
			// 	return err
			// }

			// get InventoryMaterial that availableqty > 0 and not out of stock order by date created asc limit by sum of available == reserve
			var inventoryMaterials []models.InventoryMaterial
			if err := tx.
				Where("material_id = ? AND inventory_id = ? AND availabel_qty > 0 AND is_out_of_stock = false", bom.MaterialID, mainInventory).
				Order("created_at asc").
				Find(&inventoryMaterials).Error; err != nil {
				return err
			}

			target := bom.Quantity * order.ProducedQuantity
			totalReserve := int64(0)
			var isFullFilled bool
			for _, mat := range inventoryMaterials {
				if isFullFilled {
					break
				}
				// calculate target and available qty
				requiredQty := target - totalReserve
				availabelQty := mat.AvailabelQty
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

				// update inventory material and out of stock
				mat.Reserve = updatedReserve
				mat.AvailabelQty -= rQty
				mat.IsOutOfStock = isInventoryOutOfStock
				log.Println("isInventoryOutOfStock", isInventoryOutOfStock)
				if err := tx.Save(&mat).Error; err != nil {
					return err
				}
			}

			var orderBom models.OrderBom
			orderBom.OrderID = order.ID
			orderBom.BomID = bom.ID
			orderBom.TargetQty = target
			orderBom.ReservedQty = totalReserve
			orderBom.WithdrawedQty = 0
			orderBom.IsCompletelyWithdraw = false
			orderBom.IsFullFilled = isFullFilled
			if err := tx.Create(&orderBom).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Order"})
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
