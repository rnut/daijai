package models

import "gorm.io/gorm"

type WithdrawalAdminTransaction struct {
	gorm.Model
	WithdrawalApprovementID uint
	MaterialID              uint
	Quantity                int64
	Material                *Material
}
