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
	Project          *Project
	CreatedByID      uint   `gorm:"not null"`
	CreatedBy        Member `gorm:"foreignkey:CreatedByID"`
	OrderBOMs        *[]OrderBOM
	Status           string `gorm:"default:'idle'"`
	PlanStatus       string `gorm:"default:'none'"`
	OrderReservings  *[]OrderReserving
	IsFG             bool `gorm:"default:false"`
}

const (
	OrderStatus_Idle       = "idle"
	OrderStatus_Pending    = "pending"
	OrderStatus_InProgress = "in-progress"
	OrderStatus_Done       = "done"
)
const (
	OrderPlanStatus_None     = "none"
	OrderPlanStatus_Partial  = "partial"
	OrderPlanStatus_Staged   = "staged"
	OrderPlanStatus_Complete = "complete"
)
