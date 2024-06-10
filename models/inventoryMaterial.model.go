package models

import "gorm.io/gorm"

type InventoryMaterial struct {
	gorm.Model
	MaterialID            uint `grom:"not null"`
	InventoryID           uint `grom:"not null"`
	ReceiptID             *uint
	AdjustmentID          *uint
	TransferMaterialID    *uint
	Quantity              int64
	Reserve               int64
	Withdrawed            int64
	AvailableQty          int64
	Price                 int64
	IsOutOfStock          bool
	Material              *Material         `gorm:"foreignKey:MaterialID;references:ID"`
	Inventory             *Inventory        `gorm:"foreignKey:InventoryID;references:ID"`
	Receipt               *Receipt          `gorm:"foreignKey:ReceiptID;references:ID"`
	Adjustment            *Adjustment       `gorm:"foreignKey:AdjustmentID;references:ID"`
	TransferMaterial      *TransferMaterial `gorm:"foreignKey:TransferMaterialID;references:ID"`
	Transactions          *[]InventoryMaterialTransaction
	InventoryMaterialType string
}

const (
	InventoryMaterialType_Receipt  = "receipt"
	InventoryMaterialType_Adjust   = "adjust"
	InventoryMaterialType_Transfer = "transfer"
)
