package models
import "gorm.io/gorm"

type WithdrawalTransaction struct {
	gorm.Model
	WithdrawalID uint
	OrderBomID   uint
	Quantity     int64
	Status	   string // "in-progress", "approved", "rejected"
	Withdrawal   Withdrawal `gorm:"foreignKey:WithdrawalID"`
	OrderBom     OrderBom `gorm:"foreignKey:OrderBomID"`
}

const (
	WithdrawalTransactionStatus_InProgress = "in-progress"
	WithdrawalTransactionStatus_Approved = "approved"
	WithdrawalTransactionStatus_Rejected = "rejected"
)