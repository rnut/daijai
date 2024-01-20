package models

import "gorm.io/gorm"

type Purchase struct {
	gorm.Model
	Slug              string `gorm:"unique"`
	PORefs            []PORef `gorm:"many2many:purchase_po_refs;"`
	Notes             string
	IsApprove         bool
	PurchaseMaterials []PurchaseMaterial
	CreatedByID       uint   `gorm:"not null"`
	CreatedBy         Member `gorm:"foreignkey:CreatedByID"`
}

type PORef struct {
	gorm.Model
	Slug string `gorm:"unique"`
	Purchases []Purchase `gorm:"many2many:purchase_po_refs;"`
}