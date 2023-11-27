package models

import (
	"gorm.io/gorm"
)

type PurchaseMaterial struct {
	gorm.Model
	PurchaseID uint
	MaterialID uint
	Quantity   int64
	Material   Material
}
