package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Slug      string `gorm:"unique"`
	Title     string
	Subtitle  string
	Materials []Material
	IsFG      bool `gorm:"default:false"`
}
