package models

import "gorm.io/gorm"

type Project struct {
	gorm.Model
	Slug        string
	Title       string
	Subtitle    string
	Description string
}
