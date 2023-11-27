package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Slug     string
	Title    string
	Subtitle string
}
