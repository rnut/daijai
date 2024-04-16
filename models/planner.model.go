package models

type PlanModel struct {
	OrderBOMID uint
	Quantity   int64
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
