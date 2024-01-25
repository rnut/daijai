package controllers

import (
	"daijai/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NotificaionController struct {
	DB *gorm.DB
	BaseController
}

func NewNotificationController(db *gorm.DB) *NotificaionController {
	return &NotificaionController{
		DB: db,
	}
}

func (nc *NotificaionController) GetNotifications(c *gin.Context) {
	var uid uint
	if err := nc.GetUserID(c, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var member models.Member
	if err := nc.getUserDataByUserID(nc.DB, uid, &member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var notifications []models.Notification
	result := nc.
		DB.
		Find(&notifications)
	// Where("user_id = ?", member.ID).
	// Or("topic = ?", member.Role).
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
		return
	}

	var resp struct {
		Notifications []models.Notification
		UnreadCount   int64
	}
	resp.Notifications = notifications
	resp.UnreadCount = 10
	c.JSON(http.StatusOK, resp)
}
