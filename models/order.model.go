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
	OrderStatus_Idle       = "idle"        // สร้าง
	OrderStatus_Pending    = "pending"     // จัดสรรแล้ว
	OrderStatus_InProgress = "in-progress" // กำลังเบิก
	OrderStatus_Done       = "done"        // จัดสรรและเบิกครบ
)
const (
	OrderPlanStatus_None     = "none"
	OrderPlanStatus_Partial  = "partial"  // จัดสรรไปบางส่วน แต่ยังไม่ครบ
	OrderPlanStatus_Staged   = "staged"   // จัดสรรครบแล้ว รอเบิก
	OrderPlanStatus_Complete = "complete" // จัดสรรครบแล้ว และเบิกครบ
)
