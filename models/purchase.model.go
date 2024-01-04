package models

import "gorm.io/gorm"

type Purchase struct {
	gorm.Model
	Slug              string `gorm:"unique"`
	PORef             string
	Notes             string
	IsApprove         bool
	PurchaseMaterials []PurchaseMaterial
	CreatedByID       uint   `gorm:"not null"`
	CreatedBy         Member `gorm:"foreignkey:CreatedByID"`
}
