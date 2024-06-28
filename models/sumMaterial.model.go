package models

import "gorm.io/gorm"

type SumMaterialInventory struct {
	gorm.Model
	MaterialID  uint
	InventoryID uint
	Quantity    int64
	Price       int64
}
