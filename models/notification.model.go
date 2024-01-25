package models

import "gorm.io/gorm"

type Notification struct {
	gorm.Model
	Type      string // NotificationType_*
	BadgeType string // NotificationBadgeType_*
	Title     string
	Subtitle  string
	Body      string
	Action    string // NotificationAction_*
	Icon      string
	Cover     string
	IsRead    bool
	IsSeen    bool
	Topic     string // NotificationTopic_*
	UserID    *uint
}

const (
	NotificationType_TOPIC = "topic"
	NotificationType_USER  = "user"
)

const (
	NotificationBadgeType_NONE    = "none"
	NotificationBadgeType_INFO    = "info"
	NotificationBadgeType_ERROR   = "error"
	NotificationBadgeType_WARN    = "warn"
	NotificationBadgeType_SUCCESS = "success"
)

const (
	NotificationAction_NEW_PR              = "new_pr"
	NotificationAction_APPROVED_PR         = "approved_pr"
	NotificationAction_NEW_RECEIPT         = "new_receipt"
	NotificationAction_APPROVED_RECEIPT    = "approved_receipt"
	NotificationAction_RESTOCK             = "restock"
	NotificationAction_NEW_WITHDRAWAL      = "new_withdrawal"
	NotificationAction_APPROVED_WITHDRAWAL = "approved_withdrawal"
)

const (
	NotificationTopic_GENERAL    = "general"
	NotificationTopic_ADMIN      = "admin"
	NotificationTopic_MANAGER    = "manager"
	NotificationTopic_TECHNICIAN = "technician"
	NotificationTopic_None       = "none"
)
