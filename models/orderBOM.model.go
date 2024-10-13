package models

import "gorm.io/gorm"

type OrderBOM struct {
	gorm.Model
	OrderID              uint
	Order                *Order
	MaterialID           uint
	Material             *Material
	DrawingID            uint
	Drawing              *Drawing
	TargetQty            int64
	ReservedQty          int64
	WithdrawedQty        int64
	AdjustQty            int64
	IsFullFilled         bool // จองครบหรือไม่
	IsCompletelyWithdraw bool // เบิกครบหรือไม่
}

func (OrderBOM) TableName() string {
	return "order_boms"
}
