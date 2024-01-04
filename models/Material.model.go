package models

import "gorm.io/gorm"

type Material struct {
	gorm.Model
	ImagePath  string
	Slug       string `gorm:"unique" form:"Slug"`
	Title      string `form:"Title"`
	Subtitle   string `form:"Subtitle"`
	Min        int64  `form:"Min"`
	Max        int64  `form:"Max"`
	CategoryID uint   `json:"CategoryID" form:"CategoryID"`
	Category   Category
}
