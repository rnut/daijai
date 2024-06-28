package models

import "gorm.io/gorm"

type ExtendOrder struct {
	gorm.Model
	Slug            string `gorm:"unique"`
	ProjectID       uint
	Project         *Project `gorm:"foreignkey:ProjectID"`
	Notes           string
	CreatedByID     uint    `gorm:"not null"`
	CreatedBy       *Member `gorm:"foreignkey:CreatedByID"`
	Status          string  `gorm:"default:'pending'"`
	ExtendOrderBOMs *[]ExtendOrderBOM
}

const (
	ExtendOrderStatus_Pending    = "pending"
	ExtendOrderStatus_InProgress = "in-progress"
	ExtendOrderStatus_Done       = "done"
)
