package models

import "gorm.io/gorm"

type Withdrawal struct {
	gorm.Model
	Slug                string `gorm:"unique"`
	ProjectID           uint
	Project             Project
	DrawingID           uint
	ApprovedByID        *uint
	ApprovedBy          Member `gorm:"foreignkey:ApprovedByID"`
	Notes               string
	IsApproved          bool
	WithdrawalMaterials []WithdrawalMaterial
	CreatedByID         uint   `gorm:"not null"`
	CreatedBy           Member `gorm:"foreignkey:CreatedByID"`
}
