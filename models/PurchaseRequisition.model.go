package models

import "gorm.io/gorm"

type PurchaseRequisition struct {
	gorm.Model
	DrawingID        uint
	MaterialID       uint
	Quantity         int64
	WithdrawalUserID uint
	ApprovalUserID   uint
	IsApprove        bool
}
