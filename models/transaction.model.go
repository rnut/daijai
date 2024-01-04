package models

import "gorm.io/gorm"

type Transaction struct {
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
	Ref            string
	PONumber       string
	ReceiptID      *uint
	Receipt        *Receipt
}

/// TYPE - initial | receipt    | withdrawal    | reserve  | return    | adjustment
/// REF -   null   | receipt_id | withdrawal_id | order_id | return_id | adjustment_id
