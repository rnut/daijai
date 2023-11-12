package models

import "gorm.io/gorm"

type MaterialReceipt struct {
	gorm.Model
	MaterialID    uint
	Quantity      int64
	ReceiptUserID uint
	PRID          string
}
