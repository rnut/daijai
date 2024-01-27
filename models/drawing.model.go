package models

import "gorm.io/gorm"

type Drawing struct {
	gorm.Model
	ImagePath   string
	Slug        string `form:"Slug" gorm:"unique"`
	PartNumber  string `form:"PartNumber"`
	Boms        []Bom  `form:"Bom"`
	CreatedByID uint   `gorm:"not null"`
	CreatedBy   Member `gorm:"foreignkey:CreatedByID"`
	IsFG        bool   `gorm:"default:false"`
}
