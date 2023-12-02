package models

import "gorm.io/gorm"

type Material struct {
	gorm.Model
	ImagePath        string
	Slug             string `gorm:"unique" form:"Slug"`
	Title            string `form:"Title"`
	Subtitle         string `form:"Subtitle"`
	Price            int64  `form:"Price"`
	Quantity         int64  `form:"Quantity"`
	InUseQuantity    int64  `form:"InUseQuantity"`
	IncomingQuantity int64  `form:"IncomingQuantity"`
	Supplier         string `form:"Supplier"`
	Min              int64  `form:"Min"`
	Max              int64  `form:"Max"`
	CategoryID       uint   `json:"CategoryID" form:"CategoryID"`
	Category         Category
}
