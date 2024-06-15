package models

type OrderReserving struct {
	ID                  uint
	OrderID             uint
	OrderBOMID          uint
	ReceiptID           *uint
	InventoryMaterialID uint
	Status              string // OrderReservingStatus_Reserved, OrderReservingStatus_Withdrawed
	Quantity            int64
	AdjustedQuantity    int64
	Order               *Order             `gorm:"foreignKey:OrderID"`
	OrderBOM            *OrderBOM          `gorm:"foreignKey:OrderBOMID"`
	Receipt             *Receipt           `gorm:"foreignKey:ReceiptID"`
	InventoryMaterial   *InventoryMaterial `gorm:"foreignKey:InventoryMaterialID"`
}

const (
	OrderReservingStatus_Reserved   = "reserved"
	OrderReservingStatus_Withdrawed = "withdrawed"
)
