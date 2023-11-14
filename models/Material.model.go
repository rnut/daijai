package models

import "gorm.io/gorm"

type Material struct {
	gorm.Model
	Slug             string
	Title            string
	Subtitle         string
	Price            int64
	Quantity         int64
	InUseQuantity    int64
	IncomingQuantity int64
	Supplier         string
	Min              int64
	Max              int64
}
