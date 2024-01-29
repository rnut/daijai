package controllers

import (
	"daijai/models"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReceiptController struct {
	DB *gorm.DB
	BaseController
}

func NewReceipt(db *gorm.DB) *ReceiptController {
	return &ReceiptController{
		DB: db,
	}
}

func (rc *ReceiptController) GetNewReceiptInfo(c *gin.Context) {
	var response struct {
		Slug       string
		Categories []models.Category
	}

	var categories []models.Category
	if err := rc.DB.
		Preload("Materials").
		Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Categories"})
		return
	}

	response.Categories = categories

	var slug string
	if err := rc.RequestSlug(&slug, rc.DB, "receipts"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Slug", "detail": err.Error()})
		return
	}
	response.Slug = slug

	c.JSON(http.StatusOK, response)
}

func (rc *ReceiptController) GetEditReceiptInfo(c *gin.Context) {
	var response struct {
		Recipt     models.Receipt
		Categories []models.Category
	}

	var categories []models.Category
	if err := rc.DB.
		Preload("Materials").
		Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Categories"})
		return
	}

	response.Categories = categories

	slug := c.Param("slug")
	var receipt models.Receipt
	if err := rc.DB.
		Preload("ReceiptMaterials.Material").
		Preload("ReceiptMaterials.Material.Category").
		Where("slug = ?", slug).
		First(&receipt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	response.Recipt = receipt

	c.JSON(http.StatusOK, response)
}

// CreateReceipt creates a new Receipt entry.
func (rc *ReceiptController) CreateReceipt(c *gin.Context) {
	var uid uint
	if err := rc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := rc.getUserDataByUserID(rc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var request models.Receipt
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request.CreatedByID = member.ID
	if err := rc.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&request).Error; err != nil {
			return err
		}

		title := fmt.Sprintf("Receipt was create by %s", member.FullName)
		subtitle := fmt.Sprintf("please check PR %s to see more details", request.Slug)
		notif := models.Notification{
			Type:      models.NotificationType_TOPIC,
			BadgeType: models.NotificationBadgeType_INFO,
			Title:     title,
			Subtitle:  subtitle,
			Body:      request.Slug,
			Action:    models.NotificationAction_NEW_RECEIPT,
			Icon:      "https://i.imgur.com/R3uJ7BF.png",
			Cover:     "https://i.imgur.com/R3uJ7BF.png",
			IsRead:    false,
			IsSeen:    false,
			Topic:     models.NotificationTopic_MANAGER,
		}
		if err := tx.Create(&notif).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return err
		}

		return nil
	}); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Receipt"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Receipt created successfully", "receipt": request})
}

func (rc *ReceiptController) ApproveReceipt(c *gin.Context) {
	var uid uint
	if err := rc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := rc.getUserDataByUserID(rc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	var receipt models.Receipt
	if err := rc.
		DB.
		Preload("ReceiptMaterials").
		Preload("CreatedBy").
		First(&receipt, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}
	if err := rc.DB.Transaction(func(tx *gorm.DB) error {
		// create PORef
		log.Println("create PORef:")
		log.Println(receipt.PORefNumber)
		var poRef models.PORef
		poRef.Slug = receipt.PORefNumber
		if err := tx.Create(&poRef).Error; err != nil {
			return err
		}

		var matIDs []uint
		matQuantity := make(map[uint]int64)
		matInventoryId := make(map[uint]uint)
		for _, v := range receipt.ReceiptMaterials {
			matIDs = append(matIDs, v.MaterialID)
			// update receipt material
			v.IsApproved = true
			if err := tx.Save(&v).Error; err != nil {
				return err
			}

			// create inventory material
			var inventoryMaterial models.InventoryMaterial
			inventoryMaterial.MaterialID = v.MaterialID
			inventoryMaterial.InventoryID = receipt.InventoryID
			inventoryMaterial.ReceiptID = receipt.ID
			inventoryMaterial.Quantity = v.Quantity
			inventoryMaterial.Reserve = 0
			inventoryMaterial.AvailabelQty = v.Quantity
			inventoryMaterial.IsOutOfStock = false
			inventoryMaterial.Price = v.Price
			if err := tx.Save(&inventoryMaterial).Error; err != nil {
				return err
			}

			// assign hashmap
			matQuantity[v.MaterialID] = v.Quantity
			matInventoryId[v.MaterialID] = inventoryMaterial.ID

			// create inventory materail transaction
			inventoryMaterialTransaction := models.InventoryMaterialTransaction{
				InventoryMaterialID:      inventoryMaterial.ID,
				Quantity:                 inventoryMaterial.Quantity,
				InventoryType:            models.InventoryType_INCOMING,
				InventoryTypeDescription: models.InventoryTypeDescription_INCOMINGRECEIPT,
				ExistingQuantity:         0,
				ExistingReserve:          0,
				UpdatedQuantity:          inventoryMaterial.Quantity,
				UpdatedReserve:           0,
				ReceiptID:                &receipt.ID,
			}
			if err := tx.Save(&inventoryMaterialTransaction).Error; err != nil {
				return err
			}
		}

		// get waiting material order boms
		var orderBoms []models.OrderBom
		withdrawStatuses := []string{models.OrderStatus_Waiting, models.OrderStatus_InProgress}
		if err := rc.
			DB.
			Joins("Bom").
			Joins("Order").
			Preload("Bom.Material").
			Where("is_full_filled = ?", false).
			Where("withdraw_status IN (?)", withdrawStatuses).
			Where("material_id IN ?", matIDs).
			Find(&orderBoms).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch OrderBoms"})
			return err
		}

		for _, v := range orderBoms {
			log.Println(v.ID)
			log.Println(v.Order.Slug)

		}

		var filledOrderBomIDs []uint
		// loop order boms
		for _, orderBom := range orderBoms {
			target := orderBom.TargetQty - (orderBom.ReservedQty + orderBom.WithdrawedQty)
			matID := orderBom.Bom.MaterialID
			qouta := matQuantity[matID]

			if qouta == 0 {
				continue
			}

			var quantity int64
			if qouta > target {
				quantity = target
			} else {
				quantity = qouta
			}

			inventoryMaterialID := matInventoryId[matID]

			// create order reserving
			orderReserving := models.OrderReserving{
				OrderID:             orderBom.OrderID,
				OrderBomID:          orderBom.ID,
				ReceiptID:           receipt.ID,
				InventoryMaterialID: inventoryMaterialID,
				Quantity:            quantity,
				Status:              models.OrderReservingStatus_Reserved,
			}
			if err := tx.Save(&orderReserving).Error; err != nil {
				return err
			}

			// update order bom
			orderBom.ReservedQty += quantity

			totalQty := orderBom.ReservedQty + orderBom.WithdrawedQty
			if totalQty == orderBom.TargetQty {
				orderBom.IsFullFilled = true
				filledOrderBomIDs = append(filledOrderBomIDs, orderBom.ID)
			}
			if err := tx.Save(&orderBom).Error; err != nil {
				return err
			}

			// update inventory material
			var inventoryMaterial models.InventoryMaterial
			if err := tx.First(&inventoryMaterial, inventoryMaterialID).Error; err != nil {
				return err
			}
			inventoryMaterial.Reserve += quantity
			inventoryMaterial.AvailabelQty -= quantity
			if inventoryMaterial.AvailabelQty == 0 {
				inventoryMaterial.IsOutOfStock = true
			}
			if err := tx.Save(&inventoryMaterial).Error; err != nil {
				return err
			}

			// create inventory material transaction
			ivmtReserve := models.InventoryMaterialTransaction{
				InventoryMaterialID:      inventoryMaterialID,
				Quantity:                 quantity,
				InventoryType:            models.InventoryType_RESERVE,
				InventoryTypeDescription: models.InventoryTypeDescription_FillFromReceipt,
				ExistingQuantity:         qouta,
				ExistingReserve:          0,
				UpdatedQuantity:          qouta,
				UpdatedReserve:           quantity,
				OrderID:                  &orderBom.OrderID,
			}
			if err := tx.Save(&ivmtReserve).Error; err != nil {
				return err
			}

			// update order status to in-progress
			orderBom.Order.WithdrawStatus = models.OrderStatus_InProgress
			if err := tx.Save(&orderBom.Order).Error; err != nil {
				return err
			}

			// update material quantity
			matQuantity[matID] -= quantity

			// create notification

			var withdrawal models.Withdrawal
			if err := tx.
				Where("order_id = ?", orderBom.OrderID).
				First(&withdrawal).Error; err != nil {
				return err
			}
			title := fmt.Sprintf("%s has been restock", orderBom.Bom.Material.Title)
			subtitle := "please check withdrawal request to see more details"
			notif := models.Notification{
				Type:      models.NotificationType_USER,
				BadgeType: models.NotificationBadgeType_INFO,
				Title:     title,
				Subtitle:  subtitle,
				Body:      withdrawal.Slug,
				Action:    models.NotificationAction_RESTOCK,
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
		}

		// update receipt status and approved by
		receipt.IsApproved = true
		receipt.ApprovedByID = &member.ID
		receipt.ApprovedBy = &member
		if err := tx.Save(&receipt).Error; err != nil {
			return err
		}

		if len(filledOrderBomIDs) > 0 {
			if err := tx.
				Model(&models.PurchaseSuggestion{}).
				Where("id IN ?", filledOrderBomIDs).
				Update("status", models.PurchaseSuggestionStatus_Done).
				Error; err != nil {
				return err
			}
		}

		title := fmt.Sprintf("Receipt was approved by %s", member.FullName)
		subtitle := fmt.Sprintf("please check receipt %s to see more details", receipt.Slug)
		notif := models.Notification{
			Type:      models.NotificationType_USER,
			BadgeType: models.NotificationBadgeType_INFO,
			Title:     title,
			Subtitle:  subtitle,
			Body:      receipt.Slug,
			Action:    models.NotificationAction_APPROVED_RECEIPT,
			Icon:      "https://i.imgur.com/R3uJ7BF.png",
			Cover:     "https://i.imgur.com/R3uJ7BF.png",
			IsRead:    false,
			IsSeen:    false,
			Topic:     models.NotificationTopic_None,
			UserID:    &receipt.CreatedByID,
		}
		if err := tx.Create(&notif).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return err
		}

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve Receipt"})
		return
	}
	c.JSON(http.StatusOK, receipt)
}

func (rc *ReceiptController) GetAllReceipts(c *gin.Context) {
	var uid uint
	if err := rc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := rc.getUserDataByUserID(rc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var receipts []models.Receipt
	q := rc.DB.
		Preload("ReceiptMaterials.Material").
		Preload("CreatedBy").
		Preload("ApprovedBy")

	if member.Role == "admin" {
		q.Find(&receipts)
	} else {
		q.Find(&receipts, "created_by_id = ?", member.ID)
	}
	if err := q.Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Receipts"})
		return
	}

	c.JSON(http.StatusOK, receipts)
}

// GET Recript by slug
func (rc *ReceiptController) GetReceiptBySlug(c *gin.Context) {
	slug := c.Param("slug")

	var receipt models.Receipt
	if err := rc.
		DB.
		Preload("ReceiptMaterials.Material.Category").
		Preload("CreatedBy").
		Preload("ApprovedBy").
		First(&receipt, "slug = ?", slug).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	// get all inventory materials by receipt id
	var inventoryMaterials []models.InventoryMaterial
	if err := rc.
		DB.
		Preload("Material").
		Preload("Inventory").
		Preload("Transactions.Order.Drawing").
		Preload("Transactions.Withdrawal.CreatedBy").
		Find(&inventoryMaterials, "receipt_id = ?", receipt.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"receipt": receipt, "inventoryMaterials": inventoryMaterials})
}

// UpdateReceipt updates a Receipt by ID.
func (rc *ReceiptController) UpdateReceipt(c *gin.Context) {
	slug := c.Param("slug")

	var receipt models.Receipt
	if err := rc.
		DB.
		Where("slug = ?", slug).
		Preload("ReceiptMaterials").
		First(&receipt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	if receipt.IsApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Receipt already approved"})
		return
	}

	var request models.Receipt
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := rc.DB.Transaction(func(tx *gorm.DB) error {
		receipt.Notes = request.Notes
		receipt.PORefNumber = request.PORefNumber
		if err := rc.DB.Save(&receipt).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Receipt"})
			return err
		}
		for _, v := range receipt.ReceiptMaterials {
			if err := rc.DB.Delete(&v).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Receipt"})
				return err
			}
		}

		for _, v := range request.ReceiptMaterials {
			receiptMaterial := models.ReceiptMaterial{
				ReceiptID:  receipt.ID,
				MaterialID: v.MaterialID,
				Quantity:   v.Quantity,
				Price:      v.Price,
				IsApproved: v.IsApproved,
			}
			if err := rc.DB.Save(&receiptMaterial).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Receipt"})
				return err
			}
		}
		return nil

	}); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Receipt"})
		return
	}

	c.JSON(http.StatusOK, request)
}

// DeleteReceipt deletes a Receipt by ID.
func (rc *ReceiptController) DeleteReceipt(c *gin.Context) {
	id := c.Param("id")

	var receipt models.Receipt
	if err := rc.DB.First(&receipt, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	if err := rc.DB.Delete(&receipt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete Receipt"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Receipt deleted successfully"})
}
