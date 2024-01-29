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
	if member.Role == "technician" {
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
	var ts []models.WithdrawalTransaction

	if err := wc.DB.Transaction(func(tx *gorm.DB) error {
		withdrawal.Slug = request.Slug
		withdrawal.OrderID = uint(request.OrderID)
		withdrawal.ProjectID = uint(request.ProjectID)
		withdrawal.Notes = request.Notes
		withdrawal.CreatedByID = member.ID

		if err := wc.DB.Create(&withdrawal).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Can not create Withdraw"})
			return err
		}

		// update order status
		order.WithdrawStatus = models.OrderWithdrawStatus_Idle
		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		for _, ob := range *order.OrderBoms {
			withdrawTransaction := models.WithdrawalTransaction{
				WithdrawalID: withdrawal.ID,
				OrderBomID:   ob.ID,
				Quantity:     ob.ReservedQty,
				Status:       models.WithdrawalTransactionStatus_InProgress,
			}

			if err := wc.DB.Create(&withdrawTransaction).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Can not create Withdraw"})
				return err
			}
			ts = append(ts, withdrawTransaction)
		}

		// create notification
		title := fmt.Sprintf("%s was created withdrawal request", member.FullName)
		subtitle := "please check withdrawal request to see more details"
		notif := models.Notification{
			Type:      models.NotificationType_TOPIC,
			BadgeType: models.NotificationBadgeType_INFO,
			Title:     title,
			Subtitle:  subtitle,
			Body:      withdrawal.Slug,
			Action:    models.NotificationAction_NEW_WITHDRAWAL,
			Icon:      "https://i.imgur.com/R3uJ7BF.png",
			Cover:     "https://i.imgur.com/R3uJ7BF.png",
			IsRead:    false,
			IsSeen:    false,
			Topic:     models.NotificationTopic_ADMIN,
		}
		if err := tx.Create(&notif).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return err
		}

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Withdraw"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"withdrawal": withdrawal, "transactions": ts})
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

	withdrawalID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid withdrawal ID"})
		return
	}

	var withdrawal models.Withdrawal
	if err := wc.DB.
		Preload("Project").
		Preload("CreatedBy").
		Preload("ApprovedBy").
		Preload("WithdrawalTransactions", "withdrawal_transactions.status = ?", models.WithdrawalTransactionStatus_InProgress).
		First(&withdrawal, withdrawalID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal not found", "id": withdrawalID})
		return
	}

	// Check if the withdrawal is already approved
	if withdrawal.IsApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Withdrawal is already approved"})
		return
	}

	var order models.Order
	if err := wc.
		DB.
		Preload("Drawing").
		Preload("OrderBoms.Bom").
		Preload("OrderReservings").
		Preload("OrderReservings.InventoryMaterial").
		First(&order, withdrawal.OrderID).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Order"})
		return
	}

	if err := wc.DB.Transaction(func(tx *gorm.DB) error {
		// update withdraw transactions
		for _, wt := range *withdrawal.WithdrawalTransactions {
			if wt.Status == models.WithdrawalTransactionStatus_InProgress {
				wt.Status = models.WithdrawalTransactionStatus_Approved
				if err := tx.Save(&wt).Error; err != nil {
					return err
				}
			}
		}

		var isAllCompltelyWithdraw = true

		// // update order bom
		for _, ob := range *order.OrderBoms {
			ob.WithdrawedQty += ob.ReservedQty
			ob.ReservedQty = 0
			if ob.IsFullFilled {
				ob.IsCompletelyWithdraw = true
			} else {
				isAllCompltelyWithdraw = false
			}

			if err := tx.Save(&ob).Error; err != nil {
				return err
			}
		}

		// update order reservings
		for _, reserve := range *order.OrderReservings {
			if reserve.Status == models.OrderReservingStatus_Reserved {
				reserve.Status = models.OrderReservingStatus_Withdrawed
				if err := tx.Save(&reserve).Error; err != nil {
					return err
				}
			}

			// create material transaction
			matTr := models.InventoryMaterialTransaction{
				InventoryMaterialID:      reserve.InventoryMaterialID,
				Quantity:                 reserve.Quantity,
				InventoryType:            models.InventoryType_OUTGOING,
				InventoryTypeDescription: models.InventoryTypeDescription_WITHDRAWAL,
				ExistingQuantity:         reserve.InventoryMaterial.Quantity,
				ExistingReserve:          reserve.InventoryMaterial.Reserve,
				UpdatedQuantity:          reserve.Quantity - reserve.InventoryMaterial.Quantity,
				UpdatedReserve:           reserve.InventoryMaterial.Reserve,
				WithdrawalID:             &withdrawal.ID,
			}
			if err := wc.DB.Create(&matTr).Error; err != nil {
				return err
			}

			// update inventory material
			reserve.InventoryMaterial.Withdrawed += reserve.Quantity
			if err := tx.Save(&reserve.InventoryMaterial).Error; err != nil {
				return err
			}

		}

		// update order
		if isAllCompltelyWithdraw {
			order.WithdrawStatus = models.OrderWithdrawStatus_Complete
		} else {
			order.WithdrawStatus = models.OrderWithdrawStatus_Partial
		}

		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		// create notification
		title := fmt.Sprintf("Withdrawal %s has been approved", withdrawal.Slug)
		subtitle := "please check withdrawal request to see more details"
		notif := models.Notification{
			Type:      models.NotificationType_USER,
			BadgeType: models.NotificationBadgeType_INFO,
			Title:     title,
			Subtitle:  subtitle,
			Body:      withdrawal.Slug,
			Action:    models.NotificationAction_APPROVED_WITHDRAWAL,
			Icon:      "https://i.imgur.com/R3uJ7BF.png",
			Cover:     "https://i.imgur.com/R3uJ7BF.png",
			IsRead:    false,
			IsSeen:    false,
			Topic:     models.NotificationTopic_None,
			UserID:    &withdrawal.CreatedByID,
		}
		if err := tx.Create(&notif).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return err
		}

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Withdraw"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"order": order})
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
