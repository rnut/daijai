package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Slug     string
	Username string `gorm:"unique"`
	Password string
	FullName string
	Role     string
}
