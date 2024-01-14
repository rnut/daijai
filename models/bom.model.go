package models

import "gorm.io/gorm"

type Bom struct {
	gorm.Model
	ID         uint
	Quantity   int64 `form:"Quantity"`
	DrawingID  uint  `form:"DrawingID"`
	MaterialID uint  `form:"MaterialID"`
	Material   Material
}
