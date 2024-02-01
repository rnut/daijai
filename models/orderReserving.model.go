package models

type OrderReserving struct {
	ID                  uint
	OrderID             uint
	OrderBomID          uint
	ReceiptID           uint
	InventoryMaterialID uint
	Status              string // OrderReservingStatus_Reserved, OrderReservingStatus_Withdrawed
	Quantity            int64
	Order               Order             `gorm:"foreignKey:OrderID"`
	OrderBom            OrderBom          `gorm:"foreignKey:OrderBomID"`
	Receipt             Receipt           `gorm:"foreignKey:ReceiptID"`
	InventoryMaterial   InventoryMaterial `gorm:"foreignKey:InventoryMaterialID"`
}

const (
	OrderReservingStatus_Reserved   = "reserved"
	OrderReservingStatus_Withdrawed = "withdrawed"
)
