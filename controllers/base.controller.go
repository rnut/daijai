package controllers

import (
	"daijai/models"
	"daijai/token"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BaseController struct {
	DB *gorm.DB
}

func (bc *BaseController) GetUserID(c *gin.Context, userID *uint) error {
	uid, err := token.ExtractTokenID(c)
	if err != nil {
		return err
	}
	*userID = uid
	return nil
}

// Implement this method in your controller to fetch user data
func (bc *BaseController) getUserDataByUserID(db *gorm.DB, userID uint, member *models.Member) error {
	var u models.User
	if err := db.First(&u, userID).Error; err != nil {
		log.Print(err)
		return err
	}
	*member = u.UserToMember()
	return nil
}

func (bc *BaseController) RequestSlug(slug *string, db *gorm.DB, table string) error {
	var data models.Slugger
	if err := db.
		Where("table_name = ?", table).
		First(&data).
		Error; err != nil {
		log.Print(err)
		return err
	}

	// Update the slug value
	data.Value++
	if err := db.Save(&data).Error; err != nil {
		log.Print(err)
		return err
	}
	pad := "%0" + strconv.Itoa(data.Pad) + "d"
	incrementer := fmt.Sprintf(pad, data.Value)
	combinedValue := data.Prefix + incrementer
	*slug = combinedValue
	return nil
}

func (bc *BaseController) SumMaterial(db *gorm.DB, tag string, matID uint, invID uint) error {
	// count
	var counter struct {
		Quantity   int64
		Reserved   int64
		Withdrawed int64
	}
	if err := db.
		Model(&models.InventoryMaterial{}).
		Select("SUM(quantity) as quantity, SUM(reserve) as reserved, SUM(withdrawed) as withdrawed").
		Where("material_id = ?", matID).
		// Where("inventory_id = ?", invID).
		Where("is_out_of_stock = ?", false).
		Find(&counter).Error; err != nil {
		return err
	}

	// update sum material inventory
	var sumMaterialInventory models.SumMaterialInventory
	if err := db.
		Where("material_id = ?", matID).
		// Where("inventory_id = ?", invID).
		FirstOrInit(&sumMaterialInventory).Error; err != nil {
		return err
	}
	sumMaterialInventory.MaterialID = matID
	// sumMaterialInventory.InventoryID = invID
	sumMaterialInventory.Quantity = counter.Quantity
	sumMaterialInventory.Reserved = counter.Reserved
	sumMaterialInventory.Withdrawed = counter.Withdrawed
	if err := db.Save(&sumMaterialInventory).Error; err != nil {
		return err
	}

	log.Printf("tag: %s", tag)
	log.Printf("counter: %+v\n", counter)
	log.Printf("sum: %+v\n", sumMaterialInventory)
	return nil
}

func (bc *BaseController) CreateNotification(db *gorm.DB, notification *models.Notification) error {
	return nil
	// title := fmt.Sprintf("%s was created withdrawal request", member.FullName)
	// subtitle := "please check withdrawal request to see more details"
	// notif := models.Notification{
	// 	Type:      models.NotificationType_TOPIC,
	// 	BadgeType: models.NotificationBadgeType_INFO,
	// 	Title:     title,
	// 	Subtitle:  subtitle,
	// 	Body:      withdrawal.Slug,
	// 	Action:    models.NotificationAction_NEW_WITHDRAWAL,
	// 	Icon:      "https://i.imgur.com/R3uJ7BF.png",
	// 	Cover:     "https://i.imgur.com/R3uJ7BF.png",
	// 	IsRead:    false,
	// 	IsSeen:    false,
	// 	Topic:     models.NotificationTopic_ADMIN,
	// }
	// if err := tx.Create(&notif).Error; err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return err
	// }
}

func (bc *BaseController) LogErrorAndSendBadRequest(c *gin.Context, errorMessage string) {
	log.Println(errorMessage)
	c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
}
