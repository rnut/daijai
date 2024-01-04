package models

import "gorm.io/gorm"

type Receipt struct {
	gorm.Model
	Slug             string `gorm:"unique"`
	Notes            string
	ProjectID        *uint
	Project          *Project
	PONumber         string
	InvoiceID        string
	RecipientID      *uint
	Recipient        *Member
	IsApproved       bool
	Inventory        Inventory
	InventoryID      uint `gorm:"not null"`
	ReceiptMaterials []ReceiptMaterial
	CreatedByID      uint   `gorm:"not null"`
	CreatedBy        Member `gorm:"foreignkey:CreatedByID"`
	ApprovedByID     *uint
	ApprovedBy       *Member `gorm:"foreignkey:ApprovedByID"`
}
