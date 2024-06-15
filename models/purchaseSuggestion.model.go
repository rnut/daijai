package models

import "gorm.io/gorm"

type PurchaseSuggestion struct {
	gorm.Model
	OrderBOMID uint
	OrderBOM   *OrderBOM `gorm:"foreignKey:OrderBOMID"`
	PurchaseID *uint
	Purchase   *Purchase `gorm:"foreignKey:PurchaseID"`
	Status     string
}

const (
	PurchaseSuggestionStatus_Ready      = "ready"
	PurchaseSuggestionStatus_InProgress = "in-progress"
	PurchaseSuggestionStatus_Done       = "done"
)
