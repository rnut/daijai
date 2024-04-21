package models

import "gorm.io/gorm"

type Drawing struct {
	gorm.Model
	ImagePath   string
	Slug        string
	PartNumber  string
	BOMs        []BOM
	CreatedByID uint   `gorm:"not null"`
	CreatedBy   Member `gorm:"foreignkey:CreatedByID"`
	IsFG        bool   `gorm:"default:false"`
}
