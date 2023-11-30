package models

import "gorm.io/gorm"

type Withdrawal struct {
	gorm.Model
	ProjectID           *uint
	DrawingID           *uint
	WithdrawalUserID    *uint
	ApprovalUserID      *uint
	Notes               string
	IsApproved          bool
	WithdrawalMaterials []WithdrawalMaterial
	Project             Project
}
