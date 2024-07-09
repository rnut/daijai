package models

type PlanModel struct {
	Type string // PlanType_Order, PlanType_ExtendOrder
	ID   uint
}

const (
	PlanType_Order       = "order"
	PlanType_ExtendOrder = "extend"
)

type PlanCost struct {
	Material  Material
	Quantity  int64
	TotalCost int64
}

// type SumMaterialInventory struct {
// 	gorm.Model
// 	MaterialID  uint
// 	InventoryID uint
// 	Quantity    int64
// 	Reserved    int64
// 	Withdrawed  int64
// }

// const (
// 	MaterialType_FinishedGood = "fg"
// 	MaterialType_BuiltIn      = "bi"
// )

// const (
// 	MaterialType_Param = "type"
// )
