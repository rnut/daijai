package models

import "gorm.io/gorm"

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
	ReceiptID                *uint
	Receipt                  *Receipt `gorm:"foreignKey:ReceiptID;references:ID"`
	OrderID                  *uint
	Order                    *Order `gorm:"foreignKey:OrderID;references:ID"`
	WithdrawalID             *uint
	Withdrawal               *Withdrawal `gorm:"foreignKey:WithdrawalID;references:ID"`
	AdjustmentID             *uint
	Adjustment               *Adjustment `gorm:"foreignKey:AdjustmentID;references:ID"`
	TransferMaterialID       *uint
	TransferMaterial         *TransferMaterial `gorm:"foreignKey:TransferMaterialID;references:ID"`
}

const (
	InventoryType_INCOMING    = "incoming"
	InventoryType_OUTGOING    = "outgoing"
	InventoryType_RESERVE     = "reserve"
	InventoryType_RESERVEBACK = "reserveback"
	InventoryType_TRANSFER    = "transfer"
)

const (
	InventoryTypeDescription_INCOMINGRECEIPT = "receipt"
	InventoryTypeDescription_WITHDRAWAL      = "withdrawal"
	InventoryTypeDescription_ORDER           = "order"
	InventoryTypeDescription_FillFromReceipt = "fill-order-from-receipt"
	InventoryTypeDescription_RETURN          = "return"
	InventoryTypeDescription_ADJUSTMENT      = "adjustment"
	InventoryTypeDescription_TRANSFER_IN     = "transfer-in"
	InventoryTypeDescription_TRANSFER_OUT    = "transfer-out"
)
