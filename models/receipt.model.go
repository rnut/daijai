package models

import "gorm.io/gorm"

type Receipt struct {
	gorm.Model
	Slug             string `gorm:"unique"`
	ProjectID        *uint
	PORefNumber      string
	PurchaseID       *uint
	ReceiptUserID    *uint
	Notes            string
	ReceiptMaterials []ReceiptMaterial
	CreatedByID      uint   `gorm:"not null"`
	CreatedBy        Member `gorm:"foreignkey:CreatedByID"`
}
