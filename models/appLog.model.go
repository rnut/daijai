package models

import "gorm.io/gorm"

type AppLog struct {
	gorm.Model
	MaterialID     uint
	Material       Material
	InventoryID    uint
	Inventory      Inventory
	Quantity       int64
	Reserve        int64
	QuantityChange int64
	ReserveChange  int64
	TotalQuantity  int64
	TotalReserve   int64
	Price          int64
	Type           string
	Ref            string /// REF -   null   | receipt_id | withdrawal_id | order_id | return_id | adjustment_id
	PONumber       string
	ReceiptID      *uint
	Receipt        *Receipt
}

const (
	INITIAL     = "initial"
	RECEIPT     = "receipt"
	WITHDRAWAL  = "withdrawal"
	ORDER       = "order"
	RETURN      = "return"
	ADJUSTMENT  = "adjustment"
	RESERVE     = "reserve"
	RESERVEBACK = "reserveback"
)
