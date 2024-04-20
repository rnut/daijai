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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch inventories"})
		return
	}

	incompletedStatus := []string{models.OrderWithdrawStatus_Pending, models.OrderWithdrawStatus_Idle, models.OrderWithdrawStatus_Partial}
	if err := rc.DB.
		Preload("OrderBoms.Bom.Material").
		Where("withdraw_status IN (?)", incompletedStatus).
		Find(&response.IncompleteOrders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incomplete orders"})
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

	// var response []struct {
	// 	InventoryID int               `json:"inventoryID"`
	// 	Materials   []models.Material `json:"materials"`
	// }

	// for _, v := range inventoriesIDsUint {
	// 	var materials []models.Material
	// 	if err := rc.DB.
	// 		Preload("Sums", "inventory_id = ?", v).
	// 		Find(&materials).
	// 		Error; err != nil {
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch material sum"})
	// 		return
	// 	}
	// 	response = append(response, struct {
	// 		InventoryID int               `json:"inventoryID"`
	// 		Materials   []models.Material `json:"materials"`
	// 	}{InventoryID: int(v), Materials: materials})
	// }
	c.JSON(http.StatusOK, materials)
}

// create new planner
func (rc *PlannerController) CreatePlanner(c *gin.Context) {
	var req struct {
		Plans        []models.PlanModel `json:"plans"`
		InventoryIDs []int              `json:"inventoryIDs"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// get all orderBOM by mapping orderBOMID from req
	var orderBOMIDs []uint
	for _, v := range req.Plans {
		orderBOMIDs = append(orderBOMIDs, v.OrderBOMID)
	}
	var orderBOMs []models.OrderBom
	if err := rc.DB.
		Preload("Bom.Material").
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
			var orderBOM models.OrderBom
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
			if err := rc.DB.
				Preload("Material").
				Where("material_id = ?", orderBOM.Bom.MaterialID).
				Where("inventory_id IN ?", req.InventoryIDs).
				Where("is_out_of_stock = ?", false).
				Find(&inventoryMaterials).
				Error; err != nil {
				return err
			}

			// check if inventoryMaterials empty
			if len(inventoryMaterials) == 0 {
				return fmt.Errorf("InventoryMaterial with MaterialID %d not found", orderBOM.Bom.MaterialID)
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
					OrderBomID:          orderBOM.ID,
					ReceiptID:           *invMat.ReceiptID,
					InventoryMaterialID: invMat.ID,
					Quantity:            available,
					Status:              models.OrderReservingStatus_Reserved,
				}
				if err := rc.DB.Create(&odrs).Error; err != nil {
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
				if err := rc.DB.Create(&transaction).Error; err != nil {
					return err
				}

				// update inventory material
				invMat.AvailableQty -= available
				invMat.Reserve += available

				// update inventory material is out of stock
				if invMat.AvailableQty == 0 {
					invMat.IsOutOfStock = true
				}
				if err := rc.DB.Save(&invMat).Error; err != nil {
					return err
				}

				sumReservedQty += available
				requiredQty -= available
			}
			// update isFullfilled
			orderBOM.ReservedQty += sumReservedQty
			sumAllQuantity := orderBOM.ReservedQty + orderBOM.WithdrawedQty
			if sumAllQuantity == orderBOM.TargetQty {
				orderBOM.IsFullFilled = true
			}
			if err := rc.DB.Save(&orderBOM).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orderReservings": orderReserving,
	})
}
