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

	// config.DB.Migrator().DropTable(&models.User{})
	// config.DB.Migrator().DropTable(&models.Project{})
	// config.DB.Migrator().DropTable(&models.Category{})
	// config.DB.Migrator().DropTable(&models.Material{})
	config.DB.Migrator().DropTable(&models.Drawing{})
	config.DB.Migrator().DropTable(&models.Bomb{})
	// config.DB.Migrator().DropTable(&models.Receipt{})
	// config.DB.Migrator().DropTable(&models.ReceiptMaterial{})
	// config.DB.Migrator().DropTable(&models.Withdrawal{})
	// config.DB.Migrator().DropTable(&models.WithdrawalMaterial{})
	// config.DB.Migrator().DropTable(&models.Project{})
	// config.DB.Migrator().DropTable(&models.Purchase{})
	// config.DB.Migrator().DropTable(&models.PurchaseMaterial{})

	// config.DB.AutoMigrate(&models.User{})
	// config.DB.AutoMigrate(&models.Project{})
	// config.DB.AutoMigrate(&models.Category{})
	// config.DB.AutoMigrate(&models.Material{})
	config.DB.AutoMigrate(&models.Drawing{})
	config.DB.AutoMigrate(&models.Bomb{})
	// config.DB.AutoMigrate(&models.Receipt{})
	// config.DB.AutoMigrate(&models.ReceiptMaterial{})
	// config.DB.AutoMigrate(&models.Withdrawal{})
	// config.DB.AutoMigrate(&models.WithdrawalMaterial{})
	// config.DB.AutoMigrate(&models.Purchase{})
	// config.DB.AutoMigrate(&models.PurchaseMaterial{})
}
