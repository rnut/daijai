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

func (Withdrawal) GenerateSlug() Slugger {
	return Slugger{
		TableName: "withdrawals",
		Prefix:    "Bill 001/",
		Pad:       4,
		Value:     0,
	}
}

func (Purchase) GenerateSlug() Slugger {
	return Slugger{
		TableName: "purchases",
		Prefix:    "PR-",
		Pad:       7,
		Value:     0,
	}
}
func (Receipt) GenerateSlug() Slugger {
	return Slugger{
		TableName: "receipts",
		Prefix:    "RCP-",
		Pad:       7,
		Value:     0,
	}
}

func (ExtendOrder) GenerateSlug() Slugger {
	return Slugger{
		TableName: "extend_orders",
		Prefix:    "EXT-ORD-",
		Pad:       7,
		Value:     0,
	}
}

func (Drawing) GenerateSlug() Slugger {
	return Slugger{
		TableName: "drawings",
		Prefix:    "DWG-",
		Pad:       7,
		Value:     0,
	}
}
