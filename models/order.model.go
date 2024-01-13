package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	Slug             string `gorm:"unique"`
	Note             string
	ProducedQuantity int64
	DrawingID        uint
	Drawing          Drawing
	ProjectID        uint
	Project          Project
	CreatedByID      uint   `gorm:"not null"`
	CreatedBy        Member `gorm:"foreignkey:CreatedByID"`
	OrderBoms        *[]OrderBom
}

type OrderBom struct {
	gorm.Model
	OrderID              uint
	Order                Order
	BomID                uint
	Bom                  Bom
	TargetQty            int64
	ReservedQty          int64
	WithdrawedQty        int64
	IsFullFilled         bool // จองครบหรือไม่
	IsCompletelyWithdraw bool // เบิกครบหรือไม่
}
