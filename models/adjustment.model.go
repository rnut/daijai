package models

import "gorm.io/gorm"

type Adjustment struct {
	gorm.Model
	Notes        string
	Quantity     int64      `gorm:"not null"`
	PricePerUnit int64      `gorm:"not null"`
	InventoryID  uint       `gorm:"not null"`
	Inventory    *Inventory `gorm:"foreignKey:InventoryID"`
	MaterialID   uint       `gorm:"not null"`
	Material     *Material  `gorm:"foreignKey:MaterialID"`
	CreatedByID  uint       `gorm:"not null"`
	CreatedBy    Member     `gorm:"foreignkey:CreatedByID"`
}
