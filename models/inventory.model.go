package models

import "gorm.io/gorm"

type Inventory struct {
	gorm.Model
	Slug               string `gorm:"unique"`
	Title              string
	InventoryMaterials []InventoryMaterial
	Transactions       []Transaction
}

type InventoryMaterial struct {
	gorm.Model
	MaterialID   uint `grom:"not null"`
	InventoryID  uint `grom:"not null"`
	ReceiptID    uint
	Quantity     int64
	Reserve      int64
	AvailabelQty int64
	IsOutOfStock bool
	Material     Material  `gorm:"foreignKey:MaterialID;references:ID"`
	Inventory    Inventory `gorm:"foreignKey:InventoryID;references:ID"`
	Receipt      *Receipt  `gorm:"foreignKey:ReceiptID;references:ID"`
}
