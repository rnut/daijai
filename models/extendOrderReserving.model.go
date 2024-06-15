package models

type ExtendOrderReserving struct {
	ID                  uint
	ExtendOrderID       uint
	ExtendOrderBOMID    uint
	ReceiptID           *uint
	InventoryMaterialID uint
	Status              string // OrderReservingStatus_Reserved, OrderReservingStatus_Withdrawed
	Quantity            int64
	ExtendOrder         *ExtendOrder       `gorm:"foreignKey:ExtendOrderID"`
	ExtendOrderBOM      *ExtendOrderBOM    `gorm:"foreignKey:ExtendOrderBOMID"`
	Receipt             *Receipt           `gorm:"foreignKey:ReceiptID"`
	InventoryMaterial   *InventoryMaterial `gorm:"foreignKey:InventoryMaterialID"`
}
