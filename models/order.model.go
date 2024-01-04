package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	Slug             string
	Note             string
	ProducedQuantity int64
	DrawingID        uint
	Drawing          Drawing
	ProjectID        uint
	Project          Project
	CreatedByID      uint   `gorm:"not null"`
	CreatedBy        Member `gorm:"foreignkey:CreatedByID"`
}

type OrderBom struct {
	gorm.Model
	OrderID             uint
	Order               Order
	BomID               uint
	Bom                 Bom
	Quantity            int64
	Reserved            int64
	IsFullFilled        bool
	InventoryMaterialID uint
}
