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

// NewWithdrawalController creates a new instance of WithdrawalController.
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
		Preload("Order.OrderBoms.Bom.Material").
		Preload("CreatedBy").
		Preload("ApprovedBy").
		First(&withdrawal, "slug = ?", slug).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal not found"})
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
		Preload("CreatedBy").
		Preload("ApprovedBy")
	// if member.Role == "technician" {
	// 	q.Find(&withdrawals, "created_by_id = ?", member.ID)
	// } else {
	// 	q.Find(&withdrawals)
	// }

	if err := q.Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve withdrawals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"withdrawals": withdrawals})
}

// CreateWithdrawal handles the creation of a new withdrawal transaction.
func (wc *WithdrawalController) CreateWithdrawal(c *gin.Context) {
	var request struct {
		Slug      string `json:"Slug"`
		ProjectID int    `json:"ProjectID"`
		OrderID   int    `json:"OrderID"`
		Notes     string `json:"Notes"`
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
		Preload("OrderBoms").
		Preload("OrderBoms.Bom").
		Preload("OrderReservings").
		First(&order).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Order"})
		return
	}

	if order.WithdrawStatus != models.OrderWithdrawStatus_Pending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order is not ready to create withdrawal"})
		return
	}

	var withdrawal models.Withdrawal
	var withdrawalApprovement models.WithdrawalApprovement

	if err := wc.DB.Transaction(func(tx *gorm.DB) error {
		withdrawal.Slug = request.Slug
		withdrawal.OrderID = uint(request.OrderID)
		withdrawal.ProjectID = uint(request.ProjectID)
		withdrawal.Notes = request.Notes
		withdrawal.CreatedByID = member.ID
		withdrawal.WithdrawalStatus = models.WithdrawalStatus_Pending

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
				OrderReservingID:        ob.ID,
			}

			if err := tx.Create(&withdrawTransaction).Error; err != nil {
				return err
			}
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

	if member.Role != "admin" {
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
		Preload("Withdrawal.Order.OrderBoms").
		Preload("WithdrawalTransactions").
		Preload("WithdrawalTransactions.OrderReserving.InventoryMaterial").
		Preload("WithdrawalTransactions.OrderReserving.OrderBom").
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
			odb := wts.OrderReserving.OrderBom
			odb.WithdrawedQty += wts.OrderReserving.Quantity
			odb.ReservedQty -= wts.OrderReserving.Quantity
			odb.IsCompletelyWithdraw = odb.TargetQty == odb.WithdrawedQty
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
		for _, ob := range *wapm.Withdrawal.Order.OrderBoms {
			if !ob.IsCompletelyWithdraw {
				isAllCompltelyWithdraw = false
				break
			}
		}
		if isAllCompltelyWithdraw {
			withdrawal.WithdrawalStatus = models.WithdrawalStatus_Done
			order.WithdrawStatus = models.OrderWithdrawStatus_Complete
		} else {
			withdrawal.WithdrawalStatus = models.WithdrawalStatus_InProgress
			order.WithdrawStatus = models.OrderWithdrawStatus_Partial
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
	c.JSON(http.StatusOK, wapm)
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
					OrderReservingID:        ob.ID,
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

func (mc *WithdrawalController) GetNewWithdrawInfo(c *gin.Context) {
	// get projects
	var projects []models.Project
	if err := mc.DB.Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Projects"})
		return
	}
	var orders []models.Order
	if err := mc.
		DB.
		Preload("Drawing").
		Where("withdraw_status IN (?)", models.OrderWithdrawStatus_Pending).
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
