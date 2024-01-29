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
	WithdrawStatus   string `gorm:"default:'pending'"`
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
	OrderWithdrawStatus_Pending  = "pending"
	OrderWithdrawStatus_Idle     = "idle"
	OrderWithdrawStatus_Partial  = "partial"
	OrderWithdrawStatus_Complete = "complete"
)
