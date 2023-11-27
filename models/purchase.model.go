package models

import "gorm.io/gorm"

type Purchase struct {
	gorm.Model
	ProjectID         uint
	IsApprove         bool
	PurchaseMaterials []PurchaseMaterial
}
