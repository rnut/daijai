package models

import "gorm.io/gorm"

type Withdrawal struct {
	gorm.Model
	DrawingID           uint
	WithdrawalUserID    uint
	ApprovalUserID      uint
	IsApproved          bool
	WithdrawalMaterials []WithdrawalMaterial
}

type WithdrawalMaterial struct {
	gorm.Model
	WithdrawalID uint
	MaterialID   uint
	Quantity     int64
	Material     Material
}
