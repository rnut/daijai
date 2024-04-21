package models

import "gorm.io/gorm"

type BOM struct {
	gorm.Model
	ID         uint
	Quantity   int64
	DrawingID  uint
	MaterialID uint
	Material   *Material
}
