package models

import "gorm.io/gorm"

type Receipt struct {
	gorm.Model
	ProjectID        uint
	PurchaseID       uint
	ReceiptUserID    uint
	ReceiptMaterials []ReceiptMaterial
}
