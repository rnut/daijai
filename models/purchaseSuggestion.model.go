package models

import "gorm.io/gorm"

type PurchaseSuggestion struct {
	gorm.Model
	OrderBomID uint
	OrderBom   *OrderBom `gorm:"foreignKey:OrderBomID"`
	PurchaseID *uint
	Purchase   *Purchase `gorm:"foreignKey:PurchaseID"`
	Status     string
}

const (
	PurchaseSuggestionStatus_Ready      = "ready"
	PurchaseSuggestionStatus_InProgress = "in-progress"
	PurchaseSuggestionStatus_Done       = "done"
)
