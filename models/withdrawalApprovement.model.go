package models

import "gorm.io/gorm"

type WithdrawalApprovement struct {
	gorm.Model
	WithdrawalID                uint
	WithdrawalApprovementStatus string `gorm:"default:'pending'"` // "pending", "approved", "rejected"
	Withdrawal                  *Withdrawal
	ApprovedByID                *uint
	ApprovedBy                  *Member `gorm:"foreignkey:ApprovedByID"`
	WithdrawalTransactions      *[]WithdrawalTransaction
}

const (
	WithdrawalApprovementStatus_Pending  = "pending"
	WithdrawalApprovementStatus_Approved = "approved"
	WithdrawalApprovementStatus_Rejected = "rejected"
)
