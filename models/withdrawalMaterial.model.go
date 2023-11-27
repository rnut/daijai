package models

import "gorm.io/gorm"

type WithdrawalMaterial struct {
	gorm.Model
	WithdrawalID uint
	MaterialID   uint
	Quantity     int64
	Material     Material
}
