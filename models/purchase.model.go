package models

import "gorm.io/gorm"

type Purchase struct {
	gorm.Model
	Slug              string `gorm:"unique"`
	ProjectID         uint
	IsApprove         bool
	PurchaseMaterials []PurchaseMaterial
	CreatedByID       uint   `gorm:"not null"`
	CreatedBy         Member `gorm:"foreignkey:CreatedByID"`
}
