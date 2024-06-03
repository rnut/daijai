package models

import "gorm.io/gorm"

type WithdrawalTransaction struct {
	gorm.Model
	WithdrawalApprovementID uint
	OrderReservingID        *uint
	WithdrawalApprovement   *WithdrawalApprovement `gorm:"foreignKey:WithdrawalApprovementID"`
	OrderReserving          *OrderReserving        `gorm:"foreignKey:OrderReservingID"`
}
