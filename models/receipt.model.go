package models

import "gorm.io/gorm"

type Receipt struct {
	gorm.Model
	ProjectID        uint
	PORefNumber      string
	PurchaseID       uint
	ReceiptUserID    uint
	Notes            string
	ReceiptMaterials []ReceiptMaterial
}
