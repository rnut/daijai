package models

import "gorm.io/gorm"

type Drawing struct {
	gorm.Model
	ImagePath        string
	Slug             string `form:"Slug"`
	PartNumber       string `form:"PartNumber"`
	ProducedQuantity int64  `form:"ProducedQuantity"`
	Bombs            []Bomb `form:"Bombs"`
}
