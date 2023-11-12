package main

import (
	"daijai/config"
	"daijai/models"
)

func init() {
	// config.ConnectDB()
}

func main() {
	config.ConnectDB()
	config.DB.AutoMigrate(&models.Material{})
	config.DB.AutoMigrate(&models.Bomb{})
	config.DB.AutoMigrate(&models.Drawing{})
	config.DB.AutoMigrate(&models.MaterialReceipt{})
	config.DB.AutoMigrate(&models.Withdrawal{})
	config.DB.AutoMigrate(&models.WithdrawalMaterial{})
	config.DB.AutoMigrate(&models.Project{})
	config.DB.AutoMigrate(&models.PurchaseRequisition{})
}
