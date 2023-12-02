package models

import "gorm.io/gorm"

type Bomb struct {
	gorm.Model
	Quantity   int64 `form:"Quantity"`
	DrawingID  uint  `form:"DrawingID"`
	MaterialID uint  `form:"MaterialID"`
	Material   Material
}
