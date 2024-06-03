package models

import "gorm.io/gorm"

type Withdrawal struct {
	gorm.Model
	Slug                   string `gorm:"unique"`
	ProjectID              uint
	Project                *Project
	OrderID                *uint
	Order                  *Order `gorm:"foreignkey:OrderID"`
	Notes                  string
	CreatedByID            uint    `gorm:"not null"`
	CreatedBy              *Member `gorm:"foreignkey:CreatedByID"`
	WithdrawalStatus       string  `gorm:"default:'pending'"`
	WithdrawalApprovements *[]WithdrawalApprovement
}

const (
	WithdrawalStatus_InProgress = "in-progress"
	WithdrawalStatus_Done       = "done"
)

type WithdrawalMaterial struct {
	MaterialID uint
	Quantity   int64
}
