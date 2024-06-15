package controllers

import (
	"daijai/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PlannerController struct {
	DB *gorm.DB
	BaseController
}

func NewPlannerController(db *gorm.DB) *PlannerController {
	return &PlannerController{
		DB: db,
	}
}

func (rc *PlannerController) GetNewPlannerInfo(c *gin.Context) {
	var response struct {
		Inventories      []models.Inventory `json:"inventories"`
		IncompleteOrders []models.Order     `json:"incompleteOrders"`
	}

	// get all inventories
	if err := rc.DB.Find(&response.Inventories).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to fetch inventories"})
		return
	}

	incompletedStatus := []string{models.OrderStatus_Idle, models.OrderStatus_Pending, models.OrderStatus_InProgress}
	incompletePlanedStatus := []string{models.OrderPlanStatus_None, models.OrderPlanStatus_Partial}
	if err := rc.DB.
		Preload("OrderBOMs.BOM.Material").
		Where("status IN (?)", incompletedStatus).
		Where("plan_status IN (?)", incompletePlanedStatus).
		Find(&response.IncompleteOrders).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to fetch incomplete orders"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// get materials sum by inventory
func (rc *PlannerController) GetMaterialSumByInventory(c *gin.Context) {
	params := c.Query("inventoriesIDs")
	var inventoriesIDsUint []uint
	// split string inventoriesIDs by comma
	inventoriesIDs := strings.Split(params, ",")
	fmt.Println("inventoriesIDs: ", inventoriesIDs)

	for _, id := range inventoriesIDs {
		v, _ := strconv.Atoi(id)
		inventoriesIDsUint = append(inventoriesIDsUint, uint(v))
	}
	if len(inventoriesIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "inventoriesIDs is required"})
		return
	}

	var materials []models.Material
	if err := rc.DB.
		Preload("Sums", "inventory_id IN ?", inventoriesIDsUint).
		Find(&materials).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch material sum"})
		return
	}
	c.JSON(http.StatusOK, materials)
}

// create new planner
func (rc *PlannerController) CreatePlanner(c *gin.Context) {
	var req struct {
		Plans        []models.PlanModel `json:"plans"`
		ExtendPlans  []models.PlanModel `json:"extend_plans"`
		InventoryIDs []int              `json:"inventoryIDs"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var orderIds []uint

	// get all orderBOM by mapping orderBOMID from req
	var orderBOMIDs []uint
	for _, v := range req.Plans {
		orderBOMIDs = append(orderBOMIDs, v.OrderBOMID)
	}
	var orderBOMs []models.OrderBOM
	if err := rc.DB.
		Preload("BOM.Material").
		Where("id IN ?", orderBOMIDs).
		Where("is_full_filled = ?", false).
		Where("is_completely_withdraw = ?", false).
		Find(&orderBOMs).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orderBOMs"})
		return
	}

	var orderReserving []models.OrderReserving
	if err := rc.DB.Transaction(func(tx *gorm.DB) error {
		for _, v := range req.Plans {
			requiredQty := v.Quantity
			// get orderBOMs by orderBOMID
			var orderBOM models.OrderBOM
			for _, odb := range orderBOMs {
				if odb.ID == v.OrderBOMID {
					orderBOM = odb
					break
				}
			}
			// check if orderBom empty
			if orderBOM.ID == 0 {
				return fmt.Errorf("OrderBOM with ID %d not found", v.OrderBOMID)
			}

			var inventoryMaterials []models.InventoryMaterial
			if err := tx.
				Preload("Material").
				Where("material_id = ?", orderBOM.BOM.MaterialID).
				Where("inventory_id IN ?", req.InventoryIDs).
				Where("is_out_of_stock = ?", false).
				Find(&inventoryMaterials).
				Error; err != nil {
				return err
			}

			// check if inventoryMaterials empty
			if len(inventoryMaterials) == 0 {
				return fmt.Errorf("InventoryMaterial with MaterialID %d not found - InventoryID: %d", orderBOM.BOM.MaterialID, req.InventoryIDs)
			}
			var sumReservedQty int64
			for _, invMat := range inventoryMaterials {
				var available int64
				if requiredQty <= 0 {
					break
				}

				if invMat.AvailableQty >= requiredQty {
					available = requiredQty
				} else {
					available = invMat.AvailableQty
				}

				// create OrderReserving
				odrs := models.OrderReserving{
					OrderID:             orderBOM.OrderID,
					OrderBOMID:          orderBOM.ID,
					ReceiptID:           invMat.ReceiptID,
					InventoryMaterialID: invMat.ID,
					Quantity:            available,
					Status:              models.OrderReservingStatus_Reserved,
				}
				if err := tx.Create(&odrs).Error; err != nil {
					return err
				}
				orderReserving = append(orderReserving, odrs)

				// create inventory material transaction
				transaction := models.InventoryMaterialTransaction{
					InventoryMaterialID:      invMat.ID,
					Quantity:                 available,
					InventoryType:            models.InventoryType_RESERVE,
					InventoryTypeDescription: models.InventoryTypeDescription_ORDER,
					ExistingQuantity:         invMat.Quantity,
					ExistingReserve:          invMat.Reserve,
					UpdatedQuantity:          invMat.Quantity,
					UpdatedReserve:           invMat.Reserve + available,
					OrderID:                  &orderBOM.OrderID,
				}
				if err := tx.Create(&transaction).Error; err != nil {
					return err
				}

				// update inventory material
				invMat.AvailableQty -= available
				invMat.Reserve += available

				// update inventory material is out of stock
				if invMat.AvailableQty == 0 {
					invMat.IsOutOfStock = true
				}
				if err := tx.Save(&invMat).Error; err != nil {
					return err
				}

				sumReservedQty += available
				requiredQty -= available

				// sum material
				if e := rc.SumMaterial(rc.DB, "CreatePlanner", orderBOM.BOM.MaterialID, invMat.InventoryID); e != nil {
					return e
				}
			}
			// update isFullfilled
			orderBOM.ReservedQty += sumReservedQty
			sumAllQuantity := orderBOM.ReservedQty + orderBOM.WithdrawedQty
			if sumAllQuantity == orderBOM.TargetQty {
				orderBOM.IsFullFilled = true
			}
			if err := tx.Save(&orderBOM).Error; err != nil {
				return err
			}

			orderIds = append(orderIds, orderBOM.OrderID)
		}

		// get order by orderIDs
		var orders []models.Order
		if err := tx.
			Where("id IN ?", orderIds).
			Preload("OrderBOMs").
			Find(&orders).
			Error; err != nil {
			return err
		}

		for _, order := range orders {
			// check if order is completely withdraw
			isCompletelyWithdraw := true
			isFullfilled := true
			for _, orderBOM := range *order.OrderBOMs {
				if !orderBOM.IsCompletelyWithdraw {
					isCompletelyWithdraw = false
				}

				if !orderBOM.IsFullFilled {
					isFullfilled = false
				}
			}

			if isCompletelyWithdraw {
				order.Status = models.OrderStatus_Done
				order.PlanStatus = models.OrderPlanStatus_Complete
			} else {
				if order.Status == models.OrderStatus_Idle {
					order.Status = models.OrderStatus_Pending
				}
				if isFullfilled {
					order.PlanStatus = models.OrderPlanStatus_Staged
				} else {
					order.PlanStatus = models.OrderPlanStatus_Partial
				}
			}
			if err := tx.Save(&order).Error; err != nil {
				return err
			}
		}

		result := rc.planExtendOrder(&req.ExtendPlans, req.InventoryIDs, tx)

		return result
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orderReservings": orderReserving,
	})
}

// ----------------
// Extend Plan
// ----------------
func (rc *PlannerController) planExtendOrder(plans *[]models.PlanModel, inventories []int, DB *gorm.DB) error {

	if len(*plans) == 0 {
		return nil
	}
	var orderExtendIds []uint
	for _, v := range *plans {
		orderExtendIds = append(orderExtendIds, v.OrderBOMID)
	}

	var orderExtendBOMs []models.ExtendOrderBOM
	if err := DB.
		Preload("ExtendOrderBOMs.Material").
		Where("id IN ?", orderExtendIds).
		Where("is_full_filled = ?", false).
		Where("is_completely_withdraw = ?", false).
		Find(&orderExtendBOMs).
		Error; err != nil {
		return err
	}

	var extendOrderReserving []models.ExtendOrderReserving
	for _, v := range *plans {
		requiredQty := v.Quantity

		var orderExtendBOM models.ExtendOrderBOM
		for _, oeb := range orderExtendBOMs {
			if oeb.ID == v.OrderBOMID {
				orderExtendBOM = oeb
				break
			}
		}
		if orderExtendBOM.ID == 0 {
			return fmt.Errorf("ExtendOrderBOM with ID %d not found", v.OrderBOMID)
		}

		var inventoryMaterials []models.InventoryMaterial
		if err := DB.
			Preload("Material").
			Where("material_id = ?", orderExtendBOM.MaterialID).
			Where("inventory_id IN ?", *plans).
			Where("is_out_of_stock = ?", false).
			Find(&inventoryMaterials).
			Error; err != nil {
			return err
		}

		if len(inventoryMaterials) == 0 {
			return fmt.Errorf("InventoryMaterial with MaterialID %d not found - InventoryID: %d", orderExtendBOM.MaterialID, inventories)
		}

		var sumReservedQty int64
		for _, invMat := range inventoryMaterials {
			var available int64
			if requiredQty <= 0 {
				break
			}

			if invMat.AvailableQty >= requiredQty {
				available = requiredQty
			} else {
				available = invMat.AvailableQty
			}

			// create OrderReserving
			odrs := models.ExtendOrderReserving{
				ExtendOrderID:       orderExtendBOM.ExtendOrderID,
				ExtendOrderBOMID:    orderExtendBOM.ID,
				ReceiptID:           invMat.ReceiptID,
				InventoryMaterialID: invMat.ID,
				Quantity:            available,
				Status:              models.OrderReservingStatus_Reserved,
			}
			if err := DB.Create(&odrs).Error; err != nil {
				return err
			}
			extendOrderReserving = append(extendOrderReserving, odrs)

			// create inventory material transaction
			transaction := models.InventoryMaterialTransaction{
				InventoryMaterialID:      invMat.ID,
				Quantity:                 available,
				InventoryType:            models.InventoryType_RESERVE,
				InventoryTypeDescription: models.InventoryTypeDescription_ORDER,
				ExistingQuantity:         invMat.Quantity,
				ExistingReserve:          invMat.Reserve,
				UpdatedQuantity:          invMat.Quantity,
				UpdatedReserve:           invMat.Reserve + available,
				ExtendOrderID:            &orderExtendBOM.ExtendOrderID,
			}
			if err := DB.Create(&transaction).Error; err != nil {
				return err
			}

			// update inventory material
			invMat.AvailableQty -= available
			invMat.Reserve += available

			// update inventory material is out of stock
			if invMat.AvailableQty == 0 {
				invMat.IsOutOfStock = true
			}
			if err := DB.Save(&invMat).Error; err != nil {
				return err
			}

			sumReservedQty += available
			requiredQty -= available

			// sum material
			if e := rc.SumMaterial(DB, "CreatePlanner", orderExtendBOM.MaterialID, invMat.InventoryID); e != nil {
				return e
			}

		}
		// update isFullfilled
		orderExtendBOM.ReservedQty += sumReservedQty
		sumAllQuantity := orderExtendBOM.ReservedQty + orderExtendBOM.WithdrawedQty
		if sumAllQuantity == orderExtendBOM.Quantity {
			orderExtendBOM.IsFullFilled = true
		}
		if err := DB.Save(&orderExtendBOM).Error; err != nil {
			return err
		}
	}
	return nil
}
