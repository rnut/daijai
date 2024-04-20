package models

import "gorm.io/gorm"

type TransferMaterial struct {
	gorm.Model
	Notes           string
	Quantity        int64      `gorm:"not null"`
	FromInventoryID uint       `gorm:"not null"`
	FromInventory   *Inventory `gorm:"foreignKey:FromInventoryID"`
	ToInventoryID   uint       `gorm:"not null"`
	ToInventory     *Inventory `gorm:"foreignKey:ToInventoryID"`
	MaterialID      uint       `gorm:"not null"`
	Material        *Material  `gorm:"foreignKey:MaterialID"`
	CreatedByID     uint       `gorm:"not null"`
	CreatedBy       Member     `gorm:"foreignkey:CreatedByID"`
}
