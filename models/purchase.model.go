package models

import "gorm.io/gorm"

type Purchase struct {
	gorm.Model
	Slug              string `gorm:"unique"`
	Notes             string
	ProjectID         uint
	Project           Project
	IsApprove         bool
	PurchaseMaterials []PurchaseMaterial
	CreatedByID       uint   `gorm:"not null"`
	CreatedBy         Member `gorm:"foreignkey:CreatedByID"`
}
