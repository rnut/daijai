package models

import "gorm.io/gorm"

type Slugger struct {
	gorm.Model
	TableName string
	Prefix    string
	Pad       int // e.g Pad 4 = 00001
	Value     int64
}

type Slugable interface {
	GenerateSlug() Slugger
}

func (User) GenerateSlug() Slugger {
	return Slugger{
		TableName: "users",
		Prefix:    "USR-",
		Pad:       4,
		Value:     0,
	}
}

func (Order) GenerateSlug() Slugger {
	return Slugger{
		TableName: "orders",
		Prefix:    "ORD-",
		Pad:       7,
		Value:     0,
	}
}
