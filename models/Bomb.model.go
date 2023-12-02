package models

import "gorm.io/gorm"

type Bomb struct {
	gorm.Model
	Price      int64 `form:"Price"`
	Quantity   int64 `form:"Quantity"`
	DrawingID  uint  `form:"DrawingID"`
	MaterialID uint  `form:"MaterialID"`
	Material   Material
}
