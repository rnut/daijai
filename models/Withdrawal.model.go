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
