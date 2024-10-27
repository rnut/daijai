package controllers

import (
	"daijai/models"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PlannerController struct {
	DB *gorm.DB
	BaseController
	DebugController
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
		Preload("OrderBOMs.Material").
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
	var req models.InquiryPlan
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println("=======CONFIRM Plan==========")
	log.Println("req: ")
	rc.PrintJSON(req)

	var orderIDs []uint

	for _, v := range req.Orders {
		orderIDs = append(orderIDs, v.ID)
	}

	var orders []models.Order
	if err := rc.
		DB.
		Preload("OrderBOMs", func(db *gorm.DB) *gorm.DB {
			return db.Order("id DESC")
		}).
		Preload("OrderBOMs.Material", func(db *gorm.DB) *gorm.DB {
			return db.Order("materials.id DESC")
		}).
		Where("id IN ?", orderIDs).
		Find(&orders).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	materialMaps := make(map[uint]models.PlanOrder)
	for _, order := range orders {
		for _, bom := range *order.OrderBOMs {
			ord := order
			bom.Order = &ord
			if _, ok := materialMaps[bom.MaterialID]; ok {
				planBom := models.PlanBOM{
					OrderBOM: bom,
				}
				materialMaps[bom.MaterialID] = models.PlanOrder{
					MaterialID: bom.MaterialID,
					Material:   *bom.Material,
					PlanBOMs:   append(materialMaps[bom.MaterialID].PlanBOMs, planBom),
				}
			} else {
				planBom := models.PlanBOM{
					OrderBOM: bom,
				}
				materialMaps[bom.MaterialID] = models.PlanOrder{
					MaterialID: bom.MaterialID,
					Material:   *bom.Material,
					PlanBOMs:   []models.PlanBOM{planBom},
				}
			}
		}
	}

	planOrders := make([]models.PlanOrder, 0, len(materialMaps))
	for _, v := range materialMaps {
		planOrders = append(planOrders, v)
	}

	// reorder planOrders by materialID
	sort.Slice(planOrders, func(i, j int) bool {
		return planOrders[i].MaterialID < planOrders[j].MaterialID
	})

	materialIDs := make([]uint, 0, len(materialMaps))
	for k := range materialMaps {
		materialIDs = append(materialIDs, k)
	}

	var inventoryMaterials []models.InventoryMaterial
	if err := rc.DB.
		Preload("Material").
		Where("material_id IN (?)", materialIDs).
		Where("inventory_id IN ?", req.InventoryIDs).
		Where("is_out_of_stock = ?", false).
		Find(&inventoryMaterials).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch inventory materials",
		})
	}

	if err := rc.DB.Transaction(func(tx *gorm.DB) error {
		for i, v := range planOrders {
			matID := v.MaterialID
			boms := v.PlanBOMs
			var totalUsing int64
			for j, bom := range boms {
				need := bom.OrderBOM.TargetQty - (bom.OrderBOM.ReservedQty + bom.OrderBOM.WithdrawedQty)
				var using int64
				log.Println("bom id: ", bom.OrderBOM.ID, " need: ", need, " reserved: ", bom.OrderBOM.ReservedQty, " withdrawed: ", bom.OrderBOM.WithdrawedQty)
				if need <= 0 {
					continue
				} else {
					for _, inv := range inventoryMaterials {
						if inv.MaterialID == matID && inv.AvailableQty > 0 {
							var tempExistingReserve int64
							var tempUpdatedReserve int64

							tempExistingReserve = inv.Reserve
							if inv.AvailableQty >= need {
								using = need
							} else {
								using = inv.AvailableQty
							}
							totalUsing += using
							need -= using
							boms[j].OrderBOM.ReservedQty += using

							inv.AvailableQty -= using
							inv.Reserve += using
							inv.IsOutOfStock = inv.AvailableQty == 0
							if err := tx.Save(&inv).Error; err != nil {
								return err
							}
							tempUpdatedReserve = inv.Reserve
							log.Println("update using: ", using)
							log.Println("update need: ", need)
							inventoryMaterials[i].AvailableQty = inv.AvailableQty
							inventoryMaterials[i].Reserve = inv.Reserve
							inventoryMaterials[i].IsOutOfStock = inv.IsOutOfStock

							// create order reserving
							orderReserving := models.OrderReserving{
								OrderID:             bom.OrderBOM.OrderID,
								OrderBOMID:          bom.OrderBOM.ID,
								ReceiptID:           inv.ReceiptID,
								InventoryMaterialID: inv.ID,
								Quantity:            using,
								Status:              models.OrderReservingStatus_Reserved,
							}
							if err := tx.Create(&orderReserving).Error; err != nil {
								return err
							}

							// create inventory material transaction
							transaction := models.InventoryMaterialTransaction{
								InventoryMaterialID:      inv.ID,
								Quantity:                 using,
								InventoryType:            models.InventoryType_RESERVE,
								InventoryTypeDescription: models.InventoryTypeDescription_ORDER,
								ExistingQuantity:         inv.Quantity,
								ExistingReserve:          tempExistingReserve,
								UpdatedQuantity:          inv.Quantity,
								UpdatedReserve:           tempUpdatedReserve,
								OrderID:                  &bom.OrderBOM.OrderID,
							}
							if err := tx.Create(&transaction).Error; err != nil {
								return err
							}

							if need <= 0 {
								log.Println("need = 0 -> break")
								break
							}
						}
					}

					// update bom quantity
					bom.OrderBOM.ReservedQty += totalUsing
					sumPlannedQty := bom.OrderBOM.ReservedQty + bom.OrderBOM.WithdrawedQty
					if sumPlannedQty == bom.OrderBOM.TargetQty {
						bom.OrderBOM.IsFullFilled = true
					}
					log.Println("bom id: ", bom.OrderBOM.ID, "    total using: ", using, "    total reserved: ", bom.OrderBOM.ReservedQty, "    total withdrawed: ", bom.OrderBOM.WithdrawedQty, " fullFIlled: ", bom.OrderBOM.IsFullFilled)
					// save bom
					if err := tx.Save(&bom.OrderBOM).Error; err != nil {
						return err
					}
					boms[j] = bom
					log.Println("boms[j]: ")
					rc.PrintJSON(boms[j])
				}
			}
		}

		// update order status
		if err := rc.updateOrderStatus(orderIDs, nil, tx); err != nil {
			return err
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// loop through materialIDs, inventoryIDs, and sum material
	// for _, matID := range req.MaterialIDs {
	// 	for _, invID := range req.InventoryIDs {
	// 		uMatID := uint(matID)
	// 		uInvID := uint(invID)
	// 		if err := rc.SumMaterial(rc.DB, "CreatePlanner", uMatID, uInvID); err != nil {
	// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 			return
	// 		}
	// 	}
	// }
	log.Println("=======END CONFIRM Plan==========")
	c.JSON(http.StatusOK, gin.H{"message": "Planner created successfully"})
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
			order.Status = models.OrderStatus_Pending
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

func (rc *PlannerController) InquiryPlan(c *gin.Context) {
	var req models.InquiryPlan

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var orderIDs []uint
	// var extendOrderIDs []uint

	for _, v := range req.Orders {
		orderIDs = append(orderIDs, v.ID)
	}

	var orders []models.Order
	if err := rc.
		DB.
		Preload("OrderBOMs", func(db *gorm.DB) *gorm.DB {
			return db.Order("id DESC")
		}).
		Preload("OrderBOMs.Material", func(db *gorm.DB) *gorm.DB {
			return db.Order("materials.id DESC")
		}).
		Where("id IN ?", orderIDs).
		Find(&orders).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	materialMaps := make(map[uint]models.PlanOrder)

	for _, order := range orders {
		for _, bom := range *order.OrderBOMs {
			ord := order
			bom.Order = &ord
			if _, ok := materialMaps[bom.MaterialID]; ok {
				planBom := models.PlanBOM{
					OrderBOM: bom,
				}
				materialMaps[bom.MaterialID] = models.PlanOrder{
					MaterialID: bom.MaterialID,
					Material:   *bom.Material,
					PlanBOMs:   append(materialMaps[bom.MaterialID].PlanBOMs, planBom),
				}
			} else {
				planBom := models.PlanBOM{
					OrderBOM: bom,
				}
				materialMaps[bom.MaterialID] = models.PlanOrder{
					MaterialID: bom.MaterialID,
					Material:   *bom.Material,
					PlanBOMs:   []models.PlanBOM{planBom},
				}
			}
		}
	}
	planOrders := make([]models.PlanOrder, 0, len(materialMaps))
	for _, v := range materialMaps {
		planOrders = append(planOrders, v)
	}

	// reorder planOrders by materialID
	sort.Slice(planOrders, func(i, j int) bool {
		return planOrders[i].MaterialID < planOrders[j].MaterialID
	})

	materialIDs := make([]uint, 0, len(materialMaps))
	for k := range materialMaps {
		materialIDs = append(materialIDs, k)
	}

	var sumMaterials []models.PlanSumMaterial

	if err := rc.DB.
		Model(&models.InventoryMaterial{}).
		Select("material_id, SUM(quantity) as quantity, SUM(available_qty) as available_qty, SUM(reserve) as reserve").
		Where("material_id IN ?", materialIDs).
		Where("inventory_id IN ?", req.InventoryIDs).
		Where("is_out_of_stock = ?", false).
		Group("material_id").
		Find(&sumMaterials).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch inventory materials"})
		return
	}

	mapSumMaterials := make(map[uint]models.PlanSumMaterial)
	for _, v := range sumMaterials {
		mapSumMaterials[v.MaterialID] = v
	}

	for i, v := range planOrders {
		if sumMaterial, ok := mapSumMaterials[v.MaterialID]; ok {
			planOrders[i].Capability = sumMaterial.AvailableQty

			cap := sumMaterial.AvailableQty
			for j, pb := range v.PlanBOMs {
				bom := pb.OrderBOM
				if cap <= 0 {
					break
				}
				requiredQty := bom.TargetQty - (bom.ReservedQty + bom.WithdrawedQty)
				if requiredQty <= 0 {
					continue
				} else {
					var using int64
					if cap >= requiredQty {
						using = requiredQty
					} else {
						using = cap
					}
					planOrders[i].PlanBOMs[j].NewReserveQty = using
					cap -= using

					total := using + bom.WithdrawedQty
					if total == bom.TargetQty {
						planOrders[i].PlanBOMs[j].OrderBOM.IsFullFilled = true
					}
				}
			}
		}
	}
	c.JSON(http.StatusOK, planOrders)
}
