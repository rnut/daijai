package models

import "gorm.io/gorm"

type ExtendOrderBOM struct {
	gorm.Model
	ExtendOrderID        uint
	Quantity             int64
	ReservedQty          int64
	WithdrawedQty        int64
	IsFullFilled         bool // จองครบหรือไม่
	IsCompletelyWithdraw bool // เบิกครบหรือไม่
	MaterialID           uint
	Material             *Material
}
