package controllers

import (
	"daijai/models"
	"fmt"
	"log"
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

func (rc *PlannerController) GetExtendOrders(c *gin.Context) {
	incompletedStatus := []string{models.OrderStatus_Idle, models.OrderStatus_Pending, models.OrderStatus_InProgress}

	var response []models.ExtendOrder
	if err := rc.DB.
		Preload("Project").
		Preload("ExtendOrderBOMs.Material").
		Preload("CreatedBy").
		Where("status IN (?)", incompletedStatus).
		Find(&response).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ExtendOrders"})
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
		InventoryIDs []int64            `json:"inventoryIDs"`
		MaterialIDs  []int64            `json:"materialIDs"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var orderIds []uint
	var extendOrderIds []uint

	for _, v := range req.Plans {
		if v.Type == models.PlanType_Order {
			orderIds = append(orderIds, v.ID)
		} else if v.Type == models.PlanType_ExtendOrder {
			extendOrderIds = append(extendOrderIds, v.ID)
		}
	}

	var orders []models.Order
	if err := rc.DB.
		Preload("OrderBOMs.BOM.Material").
		Where("id IN ?", orderIds).
		Find(&orders).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	orderMaps := make(map[uint]models.Order)
	for _, v := range orders {
		orderMaps[v.ID] = v
	}

	var extendOrders []models.ExtendOrder
	if err := rc.DB.
		Preload("ExtendOrderBOMs").
		Where("id IN ?", extendOrderIds).
		Find(&extendOrders).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch extend orders"})
		return
	}

	extendOrderMaps := make(map[uint]models.ExtendOrder)
	for _, v := range extendOrders {
		extendOrderMaps[v.ID] = v
	}

	var inventoryMaterials []models.InventoryMaterial
	if err := rc.DB.
		Preload("Material").
		Where("material_id IN (?)", req.MaterialIDs).
		Where("inventory_id IN ?", req.InventoryIDs).
		Where("is_out_of_stock = ?", false).
		Find(&inventoryMaterials).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch inventory materials"})
	}

	if err := rc.DB.Transaction(func(tx *gorm.DB) error {
		for _, v := range req.Plans {
			switch v.Type {
			case models.PlanType_Order:
				order := orderMaps[v.ID]
				boms := order.OrderBOMs
				for _, bom := range *boms {
					if err := rc.createPlanOrder(&bom, &inventoryMaterials, tx); err != nil {
						return err
					}
				}
			case models.PlanType_ExtendOrder:
				extendOrder := extendOrderMaps[v.ID]
				boms := extendOrder.ExtendOrderBOMs
				for _, bom := range *boms {
					if err := rc.createPlanExtendOrder(&bom, &inventoryMaterials, tx); err != nil {
						return err
					}
				}
			default:
				return fmt.Errorf("invalid PlanType")
			}
		}

		rc.updateOrderStatus(orderIds, extendOrderIds, tx)

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// loop through materialIDs, inventoryIDs, and sum material
	for _, matID := range req.MaterialIDs {
		for _, invID := range req.InventoryIDs {
			uMatID := uint(matID)
			uInvID := uint(invID)
			if err := rc.SumMaterial(rc.DB, "CreatePlanner", uMatID, uInvID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Planner created successfully"})
}

func (rc *PlannerController) createPlanOrder(bom *models.OrderBOM, inventoryMaterials *[]models.InventoryMaterial, DB *gorm.DB) error {
	if bom.IsFullFilled || bom.IsCompletelyWithdraw {
		return nil
	}

	requiredQty := bom.TargetQty - (bom.ReservedQty + bom.WithdrawedQty)
	if requiredQty <= 0 {
		return nil
	}

	sumUsing := int64(0)

	for index, invMat := range *inventoryMaterials {
		log.Println("-------before----------")
		log.Println("invMat: ", invMat.ID)
		log.Println("requiredQty: ", requiredQty)
		log.Println("invMat.AvailableQty: ", invMat.AvailableQty)
		log.Println("invMat.Reserve: ", invMat.Reserve)
		log.Println("sumUsing: ", sumUsing)
		log.Println("-----------------")

		if requiredQty <= 0 {
			log.Println("---break requiredQty <= 0---")
			break
		}

		available := invMat.AvailableQty
		var usingQty int64
		if available >= requiredQty {
			usingQty = requiredQty
		} else {
			usingQty = invMat.AvailableQty
		}
		log.Println("usingQty: ", usingQty)

		if usingQty <= 0 {
			continue
		}

		// create order reserving
		orderReserving := models.OrderReserving{
			OrderID:             bom.OrderID,
			OrderBOMID:          bom.ID,
			ReceiptID:           invMat.ReceiptID,
			InventoryMaterialID: invMat.ID,
			Quantity:            usingQty,
			Status:              models.OrderReservingStatus_Reserved,
		}
		if err := DB.Create(&orderReserving).Error; err != nil {
			return err
		}

		// create inventory material transaction
		transaction := models.InventoryMaterialTransaction{
			InventoryMaterialID:      invMat.ID,
			Quantity:                 usingQty,
			InventoryType:            models.InventoryType_RESERVE,
			InventoryTypeDescription: models.InventoryTypeDescription_ORDER,
			ExistingQuantity:         invMat.Quantity,
			ExistingReserve:          invMat.Reserve,
			UpdatedQuantity:          invMat.Quantity,
			UpdatedReserve:           invMat.Reserve + usingQty,
			OrderID:                  &bom.OrderID,
		}
		if err := DB.Create(&transaction).Error; err != nil {
			return err
		}

		// update inventory material
		invMat.AvailableQty -= usingQty
		invMat.Reserve += usingQty
		invMat.IsOutOfStock = invMat.AvailableQty == 0
		if err := DB.Save(&invMat).Error; err != nil {
			return err
		}

		// update inventory material reference
		(*inventoryMaterials)[index].AvailableQty = invMat.AvailableQty
		(*inventoryMaterials)[index].Reserve = invMat.Reserve
		(*inventoryMaterials)[index].IsOutOfStock = invMat.IsOutOfStock

		// print memory address of invMat and (*inventoryMaterials)[index]
		log.Println("-------print memory address of invMat and (*inventoryMaterials)[index]-------")
		fmt.Println((*inventoryMaterials)[index])
		fmt.Println(invMat)

		// update requiredQty
		requiredQty -= usingQty
		sumUsing += usingQty

		log.Println("-------after----------")
		log.Println("invMat.AvailableQty: ", invMat.AvailableQty)
		log.Println("invMat.Reserve: ", invMat.Reserve)
		log.Println("sumUsing: ", sumUsing)
		log.Println("requiredQty: ", requiredQty)
		log.Println("-------end----------")
	}
	log.Println("---end of invmats---")

	// update isFullfilled
	bom.ReservedQty += sumUsing
	sumAllQuantity := bom.ReservedQty + bom.WithdrawedQty
	bom.IsFullFilled = sumAllQuantity == bom.TargetQty
	if err := DB.Save(bom).Error; err != nil {
		return err
	}
	return nil
}

func (rc *PlannerController) createPlanExtendOrder(bom *models.ExtendOrderBOM, inventoryMaterials *[]models.InventoryMaterial, DB *gorm.DB) error {
	if bom.IsFullFilled || bom.IsCompletelyWithdraw {
		return nil
	}

	requiredQty := bom.Quantity - (bom.ReservedQty + bom.WithdrawedQty)
	if requiredQty <= 0 {
		return nil
	}

	sumUsing := int64(0)

	for index, invMat := range *inventoryMaterials {
		if requiredQty <= 0 {
			break
		}

		available := invMat.AvailableQty
		var usingQty int64
		if available >= requiredQty {
			usingQty = requiredQty
		} else {
			usingQty = invMat.AvailableQty
		}

		if usingQty <= 0 {
			continue
		}

		// create order reserving
		orderReserving := models.ExtendOrderReserving{
			ExtendOrderID:       bom.ExtendOrderID,
			ExtendOrderBOMID:    bom.ID,
			ReceiptID:           invMat.ReceiptID,
			InventoryMaterialID: invMat.ID,
			Status:              models.OrderReservingStatus_Reserved,
			Quantity:            usingQty,
		}
		if err := DB.Create(&orderReserving).Error; err != nil {
			return err
		}

		// create inventory material transaction
		transaction := models.InventoryMaterialTransaction{
			InventoryMaterialID:      invMat.ID,
			Quantity:                 usingQty,
			InventoryType:            models.InventoryType_RESERVE,
			InventoryTypeDescription: models.InventoryTypeDescription_EXTEND_ORDER,
			ExistingQuantity:         invMat.Quantity,
			ExistingReserve:          invMat.Reserve,
			UpdatedQuantity:          invMat.Quantity,
			UpdatedReserve:           invMat.Reserve + usingQty,
			ExtendOrderID:            &bom.ExtendOrderID,
		}
		if err := DB.Create(&transaction).Error; err != nil {
			return err
		}

		// update inventory material
		invMat.AvailableQty -= usingQty
		invMat.Reserve += usingQty
		invMat.IsOutOfStock = invMat.AvailableQty == 0
		if err := DB.Save(&invMat).Error; err != nil {
			return err
		}

		// update inventory material reference
		(*inventoryMaterials)[index].AvailableQty = invMat.AvailableQty
		(*inventoryMaterials)[index].Reserve = invMat.Reserve
		(*inventoryMaterials)[index].IsOutOfStock = invMat.IsOutOfStock

		// update requiredQty
		requiredQty -= usingQty
		sumUsing += usingQty

		// update isFullfilled
		bom.ReservedQty += sumUsing
		sumAllQuantity := bom.ReservedQty + bom.WithdrawedQty
		bom.IsFullFilled = sumAllQuantity == bom.Quantity
		if err := DB.Save(bom).Error; err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (rc *PlannerController) updateOrderStatus(orderIDs []uint, extendOrderIDs []uint, TX *gorm.DB) error {
	// get order by orderIDs
	var orders []models.Order
	if err := TX.
		Where("id IN ?", orderIDs).
		Preload("OrderBOMs").
		Find(&orders).
		Error; err != nil {
		return err
	}
	for _, order := range orders {
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
		if err := TX.Save(&order).Error; err != nil {
			return err
		}
	}

	var extendOrders []models.ExtendOrder
	if err := TX.
		Where("id IN ?", extendOrderIDs).
		Preload("ExtendOrderBOMs").
		Find(&extendOrders).
		Error; err != nil {
		return err
	}
	for _, order := range extendOrders {
		isCompletelyWithdraw := true
		isFullfilled := true
		for _, orderBOM := range *order.ExtendOrderBOMs {
			if !orderBOM.IsCompletelyWithdraw {
				isCompletelyWithdraw = false
			}

			if !orderBOM.IsFullFilled {
				isFullfilled = false
			}
		}

		if isCompletelyWithdraw {
			order.Status = models.OrderStatus_Done
			order.Status = models.OrderPlanStatus_Complete
		} else {
			if order.Status == models.OrderStatus_Idle {
				order.Status = models.OrderStatus_Pending
			}
			if isFullfilled {
				order.Status = models.OrderPlanStatus_Staged
			} else {
				order.Status = models.OrderPlanStatus_Partial
			}
		}
		if err := TX.Save(&order).Error; err != nil {
			return err
		}
	}

	return nil
}
