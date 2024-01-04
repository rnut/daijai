package models

import "gorm.io/gorm"

type ReceiptMaterial struct {
	gorm.Model
	ReceiptID  uint
	MaterialID uint
	Quantity   int64
	IsApproved bool
	Material   Material
	Price      int64
}
