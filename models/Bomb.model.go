package models

import "gorm.io/gorm"

type Bomb struct {
	gorm.Model
	Quantity   int64
	DrawingID  uint
	MaterialID uint
	Material   Material
}
