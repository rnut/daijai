package models

const (
	Plan_Order       = "order"
	Plan_ExtendOrder = "extend"
)

type PlanOrder struct {
	MaterialID uint
	Material   Material
	Capability int64
	PlanBOMs   []PlanBOM
}

type PlanBOM struct {
	OrderBOM      OrderBOM
	NewReserveQty int64
}

type InquiryPlan struct {
	InventoryIDs []int64            `json:"inventoryIDs"`
	Orders       []InquiryPlanOrder `json:"orders"`
}

type InquiryPlanOrder struct {
	ID   uint   `json:"id"`
	Type string `json:"type"`
}

type PlanSumMaterial struct {
	MaterialID   uint
	Quantity     int64
	AvailableQty int64
	Reserve      int64
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
