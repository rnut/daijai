package models

import "gorm.io/gorm"

type Withdrawal struct {
	gorm.Model
	Slug         string `gorm:"unique"`
	ProjectID    uint
	Project      *Project
	OrderID      uint
	Order        *Order `gorm:"foreignkey:OrderID"`
	ApprovedByID *uint
	ApprovedBy   *Member `gorm:"foreignkey:ApprovedByID"`
	Notes        string
	IsApproved   bool
	CreatedByID  uint    `gorm:"not null"`
	CreatedBy    *Member `gorm:"foreignkey:CreatedByID"`
	WithdrawalTransactions *[]WithdrawalTransaction
}

