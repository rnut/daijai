package models

import "gorm.io/gorm"

type ProjectStore struct {
	gorm.Model
	Slug      string `gorm:"unique"`
	Title     string
	ProjectID uint
	Project   *Project
}
