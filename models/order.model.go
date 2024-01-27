package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	Slug             string `gorm:"unique"`
	Notes            string
	ProducedQuantity int64
	DrawingID        uint
	Drawing          Drawing
	ProjectID        uint
	Project          Project
	CreatedByID      uint   `gorm:"not null"`
	CreatedBy        Member `gorm:"foreignkey:CreatedByID"`
	OrderBoms        *[]OrderBom
	WithdrawStatus   string `gorm:"default:'ready'"` // ready, in-progress, complete
	OrderReservings  *[]OrderReserving
	IsFG             bool `gorm:"default:false"`
}

type OrderBom struct {
	gorm.Model
	OrderID              uint
	Order                Order
	BomID                uint
	Bom                  *Bom
	TargetQty            int64
	ReservedQty          int64
	WithdrawedQty        int64
	IsFullFilled         bool // จองครบหรือไม่
	IsCompletelyWithdraw bool // เบิกครบหรือไม่
}

const (
	OrderStatus_Ready      = "ready"
	OrderStatus_Waiting    = "wating"
	OrderStatus_InProgress = "in-progress"
	OrderStatus_Complete   = "complete"
)
