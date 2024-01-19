package models

import "gorm.io/gorm"

type Inventory struct {
	gorm.Model
	Slug               string `gorm:"unique"`
	Title              string
	InventoryMaterials []InventoryMaterial
}

type InventoryMaterial struct {
	gorm.Model
	MaterialID   uint `grom:"not null"`
	InventoryID  uint `grom:"not null"`
	ReceiptID    uint
	Quantity     int64
	Reserve      int64
	Withdrawed      int64
	AvailabelQty int64
	Price        int64
	IsOutOfStock bool
	Material     Material  `gorm:"foreignKey:MaterialID;references:ID"`
	Inventory    Inventory `gorm:"foreignKey:InventoryID;references:ID"`
	Receipt      *Receipt  `gorm:"foreignKey:ReceiptID;references:ID"`
	Transactions []InventoryMaterialTransaction
}

type InventoryMaterialTransaction struct {
	gorm.Model
	InventoryMaterialID      uint
	InventoryMaterial        *InventoryMaterial
	Quantity                 int64
	InventoryType            string
	InventoryTypeDescription string
	ExistingQuantity         int64
	ExistingReserve          int64
	UpdatedQuantity          int64
	UpdatedReserve           int64
	ReceiptID				*uint
	Receipt					*Receipt `gorm:"foreignKey:ReceiptID;references:ID"`
	OrderID                  *uint
	Order                    *Order `gorm:"foreignKey:OrderID;references:ID"`
	WithdrawalID             *uint
	Withdrawal               *Withdrawal `gorm:"foreignKey:WithdrawalID;references:ID"`
}

const (
	InventoryType_INCOMING    = "incoming"
	InventoryType_OUTGOING    = "outgoing"
	InventoryType_RESERVE     = "reserve"
	InventoryType_RESERVEBACK = "reserveback"
)

const (
	InventoryTypeDescription_INCOMINGRECEIPT = "receipt"
	InventoryTypeDescription_WITHDRAWAL      = "withdrawal"
	InventoryTypeDescription_ORDER           = "order"
	InventoryTypeDescription_FillFromReceipt   = "fill-order-from-receipt"
	InventoryTypeDescription_RETURN          = "return"
	InventoryTypeDescription_ADJUSTMENT      = "adjustment"
)
