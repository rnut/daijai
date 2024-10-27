package controllers

import (
	"daijai/models"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

// WithdrawalController handles CRUD operations for Withdrawal model.
type WithdrawalController struct {
	DB *gorm.DB
	BaseController
}

func NewWithdrawalController(db *gorm.DB) *WithdrawalController {
	return &WithdrawalController{
		DB: db,
	}
}

func (wc *WithdrawalController) GetWithdrawalBySlug(c *gin.Context) {
	slug := c.Param("slug")
	var withdrawal models.Withdrawal
	if err := wc.DB.
		Preload("Project").
		Preload("Order.Drawing").
		Preload("WithdrawalApprovements.WithdrawalTransactions.OrderReserving.OrderBOM.Material").
		Preload("WithdrawalApprovements.WithdrawalAdminTransactions.Material").
		Preload("WithdrawalApprovements.ApprovedBy").
		Preload("Order.OrderBOMs.Material").
		Preload("CreatedBy").
		First(&withdrawal, "slug = ?", slug).Error; err != nil {
		c.JSON(http.StatusPreconditionRequired, gin.H{"error": "Withdrawal not found"})
		return
	}

	c.JSON(http.StatusOK, withdrawal)
}

func (wc *WithdrawalController) GetAllWithdrawals(c *gin.Context) {
	var uid uint
	if err := wc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := wc.getUserDataByUserID(wc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var withdrawals []models.Withdrawal
	q := wc.DB.
		Preload("Project").
		Preload("Order.Drawing").
		Preload("CreatedBy")
	if member.Role == models.ROLE_Tech {
		q.Find(&withdrawals, "created_by_id = ?", member.ID)
	} else {
		q.Find(&withdrawals)
	}

	if err := q.Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve withdrawals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"withdrawals": withdrawals})
}

func (wc *WithdrawalController) CreateNonSpecificOrderWithdrawal(c *gin.Context) {
	var request struct {
		Slug           string `json:"Slug"`
		ProjectID      int    `json:"ProjectID"`
		Notes          string `json:"Notes"`
		ProjectStoreID int    `json:"ProjectStoreID"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var uid uint
	if err := wc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := wc.getUserDataByUserID(wc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var withdrawal models.Withdrawal
	var withdrawalApprovement models.WithdrawalApprovement
	if err := wc.DB.Transaction(func(tx *gorm.DB) error {
		withdrawal.Slug = request.Slug
		withdrawal.ProjectID = uint(request.ProjectID)
		withdrawal.Notes = request.Notes
		withdrawal.CreatedByID = member.ID
		withdrawal.WithdrawalStatus = models.WithdrawalStatus_InProgress
		withdrawal.ProjectStoreID = uint(request.ProjectStoreID)

		if err := tx.Create(&withdrawal).Error; err != nil {
			return err
		}

		// create WithdrawalApprovement
		withdrawalApprovement = models.WithdrawalApprovement{
			WithdrawalID:                withdrawal.ID,
			WithdrawalApprovementStatus: models.WithdrawalApprovementStatus_Pending,
		}
		if err := tx.Create(&withdrawalApprovement).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		message := fmt.Sprintf("Failed to create Withdraw: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": message})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Withdrawal created successfully", "withdrawal": withdrawal})
}

func (wc *WithdrawalController) CreateWithdrawalAdmin(c *gin.Context) {
	var request struct {
		Slug              string                      `json:"Slug"`
		ProjectID         int                         `json:"ProjectID"`
		WithdrawMaterials []models.WithdrawalMaterial `json:"WithdrawMaterials"`
		Notes             string                      `json:"Notes"`
		ProjectStoreID    int                         `json:"ProjectStoreID"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println("--------request-------")
	log.Println(request)

	var uid uint
	if err := wc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := wc.getUserDataByUserID(wc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check role is belong to admin or manager
	canFindAll := member.Role == models.ROLE_Admin
	if !canFindAll {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Permission Denied"})
		return
	}

	var withdrawal models.Withdrawal
	var withdrawalApprovement models.WithdrawalApprovement
	if err := wc.DB.Transaction(func(tx *gorm.DB) error {

		withdrawal.Slug = request.Slug
		withdrawal.ProjectID = uint(request.ProjectID)
		withdrawal.Notes = request.Notes
		withdrawal.CreatedByID = member.ID
		withdrawal.WithdrawalStatus = models.WithdrawalStatus_Done
		withdrawal.ProjectStoreID = uint(request.ProjectStoreID)

		if err := tx.Create(&withdrawal).Error; err != nil {
			return err
		}

		// create WithdrawalApprovement
		withdrawalApprovement = models.WithdrawalApprovement{
			WithdrawalID:                withdrawal.ID,
			WithdrawalApprovementStatus: models.WithdrawalApprovementStatus_Approved,
			ApprovedByID:                &member.ID,
		}
		if err := tx.Create(&withdrawalApprovement).Error; err != nil {
			return err
		}

		// create admin withdrawal transaction
		for _, wm := range request.WithdrawMaterials {
			awt := models.WithdrawalAdminTransaction{
				WithdrawalApprovementID: withdrawalApprovement.ID,
				MaterialID:              wm.MaterialID,
				Quantity:                wm.Quantity,
			}
			if err := tx.Create(&awt).Error; err != nil {
				return err
			}

			// get InventoryMaterial by id
			var invMats []models.InventoryMaterial
			if err := tx.
				Where("material_id = ?", wm.MaterialID).
				Where("available_qty > ?", 0).
				Where("is_out_of_stock = ?", false).
				Find(&invMats).
				Error; err != nil {
				return err
			}

			needQty := wm.Quantity
			for _, invMat := range invMats {
				if needQty == 0 {
					break
				}

				existingQty := invMat.AvailableQty
				if invMat.Quantity >= needQty {
					invMat.Withdrawed += needQty
					invMat.AvailableQty = invMat.AvailableQty - needQty
					needQty = 0
				} else {
					invMat.Withdrawed += invMat.Quantity
					invMat.AvailableQty = 0
					needQty -= invMat.Quantity
				}
				log.Println(invMat.AvailableQty)

				// update out of stock
				if invMat.AvailableQty == 0 {
					invMat.IsOutOfStock = true
				}

				if err := tx.Save(&invMat).Error; err != nil {
					return err
				}

				// sum material
				matID := invMat.MaterialID
				invID := invMat.InventoryID
				wc.SumMaterial(tx, "withdrawal", matID, invID)

				// create material transaction
				matTr := models.InventoryMaterialTransaction{
					InventoryMaterialID:      invMat.ID,
					Quantity:                 invMat.Quantity,
					InventoryType:            models.InventoryType_OUTGOING,
					InventoryTypeDescription: models.InventoryTypeDescription_WITHDRAWAL,
					ExistingQuantity:         existingQty,
					ExistingReserve:          invMat.Reserve,
					UpdatedQuantity:          invMat.AvailableQty,
					UpdatedReserve:           invMat.Reserve,
					WithdrawalID:             &withdrawal.ID,
				}
				if err := tx.Create(&matTr).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		message := fmt.Sprintf("Failed to create Withdraw: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": message})
		return

	}
	c.JSON(http.StatusCreated, gin.H{"message": "Withdrawal created successfully", "withdrawal": withdrawal})
}

// CreateWithdrawal handles the creation of a new withdrawal transaction.
func (wc *WithdrawalController) CreateWithdrawal(c *gin.Context) {
	var request struct {
		Slug           string `json:"Slug"`
		ProjectID      int    `json:"ProjectID"`
		OrderID        int    `json:"OrderID"`
		Notes          string `json:"Notes"`
		ProjectStoreID int    `json:"ProjectStoreID"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var uid uint
	if err := wc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := wc.getUserDataByUserID(wc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.Order
	if err := wc.DB.
		Preload("Drawing").
		Preload("OrderBOMs").
		Preload("OrderReservings").
		First(&order, request.OrderID).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Order"})
		return
	}

	var withdrawal models.Withdrawal
	var withdrawalApprovement models.WithdrawalApprovement
	orderID := uint(request.OrderID)

	if err := wc.DB.Transaction(func(tx *gorm.DB) error {
		withdrawal.Slug = request.Slug
		withdrawal.OrderID = &orderID
		withdrawal.ProjectID = uint(request.ProjectID)
		withdrawal.Notes = request.Notes
		withdrawal.CreatedByID = member.ID
		withdrawal.WithdrawalStatus = models.WithdrawalStatus_InProgress
		withdrawal.ProjectStoreID = uint(request.ProjectStoreID)

		if err := tx.Create(&withdrawal).Error; err != nil {
			return err
		}

		// create WithdrawalApprovement
		withdrawalApprovement = models.WithdrawalApprovement{
			WithdrawalID:                withdrawal.ID,
			WithdrawalApprovementStatus: models.WithdrawalApprovementStatus_Pending,
		}
		if err := tx.Create(&withdrawalApprovement).Error; err != nil {
			return err
		}

		for _, ob := range *order.OrderReservings {
			withdrawTransaction := models.WithdrawalTransaction{
				WithdrawalApprovementID: withdrawalApprovement.ID,
				OrderReservingID:        &ob.ID,
			}

			if err := tx.Create(&withdrawTransaction).Error; err != nil {
				return err
			}
		}

		// update order status
		order.Status = models.OrderStatus_InProgress
		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		// TODO: - create notification from base controller
		return nil
	}); err != nil {
		message := fmt.Sprintf("Failed to create Withdraw: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": message})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Withdrawal created successfully", "withdrawal": withdrawal})
}

func (wc *WithdrawalController) UpdateWithdrawal(c *gin.Context) {
	var request struct {
		Withdrawal struct {
			ProjectID uint
			Notes     string
		}
	}

	withdrawalID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid drawing ID"})
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var withdrawal models.Withdrawal
	if err := wc.DB.Preload("Project").Preload("CreatedBy").First(&withdrawal, withdrawalID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal not found", "id": withdrawalID})
		return
	}

	if err := wc.DB.Transaction(func(tx *gorm.DB) error {
		pID := request.Withdrawal.ProjectID
		withdrawal.ProjectID = pID
		withdrawal.Project.ID = pID
		withdrawal.Notes = request.Withdrawal.Notes
		if err := tx.Save(&withdrawal).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Update Withdraw successfully", "withdrawal": withdrawal})
}

// // ApproveWithdrawal approves a withdrawal transaction and updates the material quantity.
func (wc *WithdrawalController) ApproveWithdrawal(c *gin.Context) {

	var uid uint
	if err := wc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := wc.getUserDataByUserID(wc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	canFindAll := member.Role == models.ROLE_Admin || member.Role == models.ROLE_Manager
	if !canFindAll {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Permission Denied"})
		return
	}

	withdrawalApprovementID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid withdrawal ID"})
		return
	}

	var wapm models.WithdrawalApprovement
	if err := wc.DB.
		Preload("Withdrawal.Order.OrderBOMs").
		Preload("WithdrawalTransactions.OrderReserving.InventoryMaterial").
		Preload("WithdrawalTransactions.OrderReserving.OrderBOM").
		First(&wapm, withdrawalApprovementID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal not found", "id": withdrawalApprovementID})
		return
	}
	// Check if the withdrawal is already approved
	if wapm.WithdrawalApprovementStatus != models.WithdrawalApprovementStatus_Pending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Withdrawal is already approved or rejected"})
		return
	}

	if err := wc.DB.Transaction(func(tx *gorm.DB) error {
		// update withdraw transactions
		wapm.WithdrawalApprovementStatus = models.WithdrawalApprovementStatus_Approved
		wapm.ApprovedByID = &member.ID
		if err := tx.Save(&wapm).Error; err != nil {
			return err
		}

		for _, wts := range *wapm.WithdrawalTransactions {
			odb := wts.OrderReserving.OrderBOM
			odb.WithdrawedQty += wts.OrderReserving.Quantity
			odb.ReservedQty -= wts.OrderReserving.Quantity
			if err := tx.Save(&odb).Error; err != nil {
				return err
			}

			// update order reservings
			wts.OrderReserving.Status = models.OrderReservingStatus_Withdrawed
			if err := tx.Save(&wts.OrderReserving).Error; err != nil {
				return err
			}

			reserve := wts.OrderReserving
			// update inventory material
			reserve.InventoryMaterial.Withdrawed += reserve.Quantity
			if err := tx.Save(&reserve.InventoryMaterial).Error; err != nil {
				return err
			}

			// sum material
			matID := reserve.InventoryMaterial.MaterialID
			invID := reserve.InventoryMaterial.InventoryID
			wc.SumMaterial(tx, "withdrawal", matID, invID)

			// create material transaction
			matTr := models.InventoryMaterialTransaction{
				InventoryMaterialID:      reserve.InventoryMaterialID,
				Quantity:                 reserve.Quantity,
				InventoryType:            models.InventoryType_OUTGOING,
				InventoryTypeDescription: models.InventoryTypeDescription_WITHDRAWAL,
				ExistingQuantity:         reserve.InventoryMaterial.Quantity,
				ExistingReserve:          reserve.InventoryMaterial.Reserve,
				UpdatedQuantity:          reserve.InventoryMaterial.Quantity - reserve.Quantity,
				UpdatedReserve:           reserve.InventoryMaterial.Reserve,
				WithdrawalID:             &wapm.WithdrawalID,
			}
			if err := wc.DB.Create(&matTr).Error; err != nil {
				return err
			}
		}

		// check is all order bom is completely withdraw
		var isAllCompltelyWithdraw = true
		withdrawal := wapm.Withdrawal
		order := withdrawal.Order
		var orderBoms []models.OrderBOM
		if err := tx.
			Where("order_id = ?", order.ID).
			Find(&orderBoms).
			Error; err != nil {
			return err
		}

		for _, ob := range orderBoms {
			completely := ob.TargetQty == ob.WithdrawedQty
			ob.IsCompletelyWithdraw = completely
			if !completely {
				isAllCompltelyWithdraw = false
				break
			}

			if err := tx.Save(&ob).Error; err != nil {
				return err
			}
		}
		if isAllCompltelyWithdraw {
			withdrawal.WithdrawalStatus = models.WithdrawalStatus_Done
			order.Status = models.OrderStatus_Done
			order.PlanStatus = models.OrderPlanStatus_Complete
		} else {
			withdrawal.WithdrawalStatus = models.WithdrawalStatus_InProgress
			order.Status = models.OrderStatus_InProgress
		}
		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		if err := tx.Save(&withdrawal).Error; err != nil {
			return err
		}

		// create notification
		// title := fmt.Sprintf("Withdrawal %s has been approved", withdrawal.Slug)
		// subtitle := "please check withdrawal request to see more details"
		// Type:      models.NotificationType_USER,
		// BadgeType: models.NotificationBadgeType_INFO,
		// Body:      withdrawal.Slug,
		// Action:    models.NotificationAction_APPROVED_WITHDRAWAL,
		// Icon:      "https://i.imgur.com/R3uJ7BF.png",
		// Topic:     models.NotificationTopic_None,
		// UserID:    &withdrawal.CreatedByID,
		// TODO:  create notification

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Withdraw"})
		return
	}

	var withdrawal models.Withdrawal
	if err := wc.DB.
		Preload("Project").
		Preload("Order.Drawing").
		Preload("WithdrawalApprovements.WithdrawalTransactions.OrderReserving.OrderBOM.Material").
		Preload("WithdrawalApprovements.ApprovedBy").
		Preload("Order.OrderBOMs.Material").
		Preload("CreatedBy").
		First(&withdrawal, wapm.ID).Error; err != nil {
		c.JSON(http.StatusPreconditionRequired, gin.H{"error": "Withdrawal not found"})
		return
	}

	c.JSON(http.StatusOK, withdrawal)
}

// CreatePartialWithdrawal handles the creation of a partial withdrawal transaction.
func (wc *WithdrawalController) CreatePartialWithdrawal(c *gin.Context) {
	var request struct {
		WithdrawalID int `json:"WithdrawalID"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var uid uint
	if err := wc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := wc.getUserDataByUserID(wc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var withdrawal models.Withdrawal
	if err := wc.DB.
		Preload("Project").
		Preload("Order.OrderReservings").
		First(&withdrawal, request.WithdrawalID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal not found"})
		return
	}

	if withdrawal.CreatedByID != member.ID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Permission Denied"})
		return
	}

	if withdrawal.WithdrawalStatus == models.WithdrawalStatus_Done {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Withdrawal is already completed"})
		return
	}

	// create withdrawal approvement
	var withdrawalApprovement models.WithdrawalApprovement
	if err := wc.DB.Transaction(func(tx *gorm.DB) error {
		withdrawalApprovement = models.WithdrawalApprovement{
			WithdrawalID:                withdrawal.ID,
			WithdrawalApprovementStatus: models.WithdrawalApprovementStatus_Pending,
		}

		if err := tx.Create(&withdrawalApprovement).Error; err != nil {
			return err
		}

		for _, ob := range *withdrawal.Order.OrderReservings {
			if ob.Status == models.OrderReservingStatus_Reserved {
				// create withdrawal transaction
				withdrawTransaction := models.WithdrawalTransaction{
					WithdrawalApprovementID: withdrawalApprovement.ID,
					OrderReservingID:        &ob.ID,
				}
				if err := tx.Create(&withdrawTransaction).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Withdraw"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Withdrawal created successfully", "withdrawalApprovement": withdrawalApprovement})

}

// DeleteMaterial deletes a specific material by ID.
func (mc *WithdrawalController) DeleteWithdraw(c *gin.Context) {
	materialID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid material ID"})
		return
	}

	if err := mc.DB.Delete(&models.Withdrawal{}, materialID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete withdrawal"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal deleted successfully"})
}

func (mc *WithdrawalController) GetNewWithdrawAdminInfo(c *gin.Context) {
	// get projects
	var projects []models.Project
	if err := mc.
		DB.
		Preload("ProjectStores").
		Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Projects"})
		return
	}

	// get categories
	var categories []models.Category
	if err := mc.DB.
		Preload("Materials", func(db *gorm.DB) *gorm.DB {
			db = db.Order("id asc")
			return db
		}).
		Preload("Materials.Sums").
		Find(&categories).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Categories"})
		return
	}

	// get slug
	var slug string
	if err := mc.RequestSlug(&slug, mc.DB, "withdrawals"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Slug", "detail": err.Error()})
		return
	}

	var response struct {
		Slug       string
		Projects   []models.Project
		Categories []models.Category
	}
	response.Slug = slug
	response.Projects = projects
	response.Categories = categories
	c.JSON(http.StatusOK, response)
}

func (mc *WithdrawalController) GetNewWithdrawInfo(c *gin.Context) {
	// get projects
	var projects []models.Project
	if err := mc.
		DB.
		Preload("ProjectStores").
		Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Projects"})
		return
	}
	var orders []models.Order
	if err := mc.
		DB.
		Preload("Drawing").
		Where("status IN (?)", models.OrderStatus_Pending).
		Find(&orders).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Orders"})
		return
	}

	// get slug
	var slug string
	if err := mc.RequestSlug(&slug, mc.DB, "withdrawals"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Slug", "detail": err.Error()})
		return
	}

	var response struct {
		Slug     string
		Projects []models.Project
		Orders   []models.Order
	}
	response.Slug = slug
	response.Projects = projects
	response.Orders = orders
	c.JSON(http.StatusOK, response)
}

func (mc *WithdrawalController) AdjustOrderReserving(c *gin.Context) {
	orderReservingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order reserving ID"})
		return
	}

	var request struct {
		WithdrawalID     int64 `json:"WithdrawalID"`
		AdjustedQuantity int64 `json:"AdjustedQuantity"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var orderReserving models.OrderReserving
	if err := mc.
		DB.
		Preload("InventoryMaterial").
		First(&orderReserving, orderReservingID).
		Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "OrderReserving not found"})
		return
	}

	if orderReserving.Status == models.OrderReservingStatus_Withdrawed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OrderReserving is already withdrawn"})
		return
	}

	// create db transaction
	if err := mc.DB.Transaction(func(tx *gorm.DB) error {
		diff := request.AdjustedQuantity - orderReserving.Quantity
		needQty := diff
		log.Printf("diff: %d\n", diff)
		if orderReserving.InventoryMaterial.AvailableQty > 0 {
			ivtm := orderReserving.InventoryMaterial
			avialableQty := ivtm.AvailableQty
			log.Printf("--------use same ivtm--------\n %d\n %+v\n", avialableQty, ivtm)

			var used int64
			if avialableQty >= diff {
				log.Printf("use only one ivtm: avialableQty >= diff\n")
				used = diff
			} else {
				log.Printf("cut off same ivtm\n")
				used = avialableQty
			}

			if err := createMaterialTransaction(ivtm, used, orderReserving.OrderID, tx); err != nil {
				return err
			}

			ivtm.Reserve += used
			ivtm.AvailableQty -= used
			ivtm.IsOutOfStock = ivtm.AvailableQty == 0
			if err := tx.Save(&ivtm).Error; err != nil {
				return err
			}

			// update order reservings
			orderReserving.Quantity += used
			// calculate needQty
			needQty = diff - used
		}

		log.Printf("needQty: %d", needQty)
		if needQty != 0 {
			log.Printf("--------new ivtm--------\n needQty: %d", needQty)
			// find available inventory material
			var invMats []models.InventoryMaterial
			if err := tx.
				Where("material_id = ?", orderReserving.InventoryMaterial.MaterialID).
				Where("available_qty > ?", 0).
				Where("is_out_of_stock = ?", false).
				Find(&invMats).
				Error; err != nil {
				return err
			}
			log.Printf("--------invMats--------\n%+v\n count: %d", invMats, len(invMats))
			log.Printf("--------for loop--------\n")
			for _, invMat := range invMats {
				// existingReserve := invMat.Reserve
				var used int64
				if needQty == 0 {
					break
				}
				log.Printf("invMat:%+v\n needQty: %d", invMat, needQty)

				if invMat.AvailableQty >= needQty {
					used = needQty
					needQty = 0
				} else {
					used = invMat.AvailableQty
					needQty -= invMat.Quantity
				}

				if err := createMaterialTransaction(&invMat, used, orderReserving.OrderID, tx); err != nil {
					return err
				}

				invMat.Reserve += used
				invMat.AvailableQty -= used
				invMat.IsOutOfStock = invMat.AvailableQty == 0
				if err := tx.Save(&invMat).Error; err != nil {
					return err
				}

				// update order reservings
				orderReserving.Quantity += used
			}
		}

		// save order reserving
		if err := tx.Save(&orderReserving).Error; err != nil {
			return err
		}

		// sum inventory material
		matID := orderReserving.InventoryMaterial.MaterialID
		invID := orderReserving.InventoryMaterial.InventoryID
		mc.SumMaterial(tx, "adjust-withdrawal", matID, invID)

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to adjust order reserving"})
		return
	}
	c.JSON(http.StatusOK, orderReserving)
}

func createMaterialTransaction(invMat *models.InventoryMaterial, reserveQty int64, orderID uint, tx *gorm.DB) error {
	matTr := models.InventoryMaterialTransaction{
		InventoryMaterialID:      invMat.ID,
		Quantity:                 invMat.Quantity,
		InventoryType:            models.InventoryType_RESERVE,
		InventoryTypeDescription: models.InventoryTypeDescription_ORDER,
		ExistingQuantity:         invMat.Quantity,
		ExistingReserve:          invMat.Reserve,
		UpdatedQuantity:          invMat.Quantity,
		UpdatedReserve:           invMat.Reserve + reserveQty,
		OrderID:                  &orderID,
	}
	if err := tx.Create(&matTr).Error; err != nil {
		return err
	} else {
		return nil
	}
}
