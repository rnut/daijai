package models

import "gorm.io/gorm"

type Drawing struct {
	gorm.Model
	Slug             string
	PartNumber       string
	ProducedQuantity int64
	Bombs            []Bomb
}
