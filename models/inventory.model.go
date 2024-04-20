package models

import "gorm.io/gorm"

type Inventory struct {
	gorm.Model
	Slug               string `gorm:"unique"`
	Title              string
	InventoryMaterials []InventoryMaterial
}
