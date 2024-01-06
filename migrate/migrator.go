package main

import (
	"daijai/config"
	"daijai/models"

	"gorm.io/gorm"
)

func init() {
	// config.ConnectDB()
}

func main() {
	config.ConnectDB()

	// config.DB.Migrator().DropTable(&models.User{})
	// config.DB.Migrator().DropTable(&models.Material{})
	// config.DB.Migrator().DropTable(&models.Project{})
	// config.DB.Migrator().DropTable(&models.Category{})
	// config.DB.Migrator().DropTable(&models.Inventory{})

	// config.DB.Migrator().DropTable(&models.Drawing{})
	// config.DB.Migrator().DropTable(&models.Bom{})
	// config.DB.Migrator().DropTable(&models.Withdrawal{})
	// config.DB.Migrator().DropTable(&models.WithdrawalMaterial{})
	// config.DB.Migrator().DropTable(&models.Project{})
	// config.DB.Migrator().DropTable(&models.Purchase{})
	// config.DB.Migrator().DropTable(&models.PurchaseMaterial{})

	config.DB.Migrator().DropTable(&models.Receipt{})
	config.DB.Migrator().DropTable(&models.ReceiptMaterial{})
	config.DB.Migrator().DropTable(&models.AppLog{})
	config.DB.Migrator().DropTable(&models.InventoryMaterial{})
	config.DB.Migrator().DropTable(&models.Order{})
	config.DB.Migrator().DropTable(&models.OrderBom{})
	config.DB.Migrator().DropTable(&models.InventoryMaterialTransaction{})

	// config.DB.AutoMigrate(&models.User{})
	// config.DB.AutoMigrate(&models.Project{})
	// config.DB.AutoMigrate(&models.Inventory{})
	// config.DB.AutoMigrate(&models.Category{})
	// config.DB.AutoMigrate(&models.Material{})
	// config.DB.AutoMigrate(&models.Drawing{})
	// config.DB.AutoMigrate(&models.Bom{})
	// config.DB.AutoMigrate(&models.Withdrawal{})
	// config.DB.AutoMigrate(&models.WithdrawalMaterial{})
	// config.DB.AutoMigrate(&models.Purchase{})
	// config.DB.AutoMigrate(&models.PurchaseMaterial{})

	config.DB.AutoMigrate(&models.Receipt{})
	config.DB.AutoMigrate(&models.ReceiptMaterial{})
	config.DB.AutoMigrate(&models.AppLog{})
	config.DB.AutoMigrate(&models.InventoryMaterial{})
	config.DB.AutoMigrate(&models.Order{})
	config.DB.AutoMigrate(&models.OrderBom{})
	config.DB.AutoMigrate(&models.InventoryMaterialTransaction{})

	// initAdmin(config.DB)
	// initUser(config.DB)
	// initCategory(config.DB)
	// initInventory(config.DB)
	// initProject(config.DB)
	// initMaterial(config.DB)
}

func initAdmin(db *gorm.DB) {
	admin := models.User{
		Slug:     "admin-001",
		Username: "salah",
		Password: "$2a$10$zeswe0/DbG/2k.4KlIbLTO2bYwmvpbXMYp2aJf.dyy7FXyHOmg9xm",
		FullName: "John Doe",
		Role:     "admin",
		Tel:      "0990938983",
	}

	db.Create(&admin)
}

func initUser(db *gorm.DB) {
	user := models.User{
		Slug:     "user-001",
		Username: "greliss",
		Password: "$2a$10$zeswe0/DbG/2k.4KlIbLTO2bYwmvpbXMYp2aJf.dyy7FXyHOmg9xm",
		FullName: "Jack Grelish",
		Role:     "user",
		Tel:      "0992221111",
	}

	db.Create(&user)
}

func initCategory(db *gorm.DB) {
	item := models.Category{
		Slug:     "PremierLeague",
		Title:    "พรีเมียร์ลีก",
		Subtitle: "ลีคสูงสุดประเทศอังกฤษ",
	}
	db.Create(&item)

	item2 := models.Category{
		Slug:     "Series A",
		Title:    "galcao series R",
		Subtitle: "ลีคสูงสุดประเทศอิตาลี",
	}

	db.Create(&item2)
}

func initInventory(db *gorm.DB) {
	inventory := models.Inventory{
		Slug:  "inventory-001",
		Title: "Main Inventory",
	}

	inventory2 := models.Inventory{
		Slug:  "inventory-002",
		Title: "Factory Inventory",
	}
	db.Create(&inventory)
	db.Create(&inventory2)
}

// init project model
func initProject(db *gorm.DB) {
	project := models.Project{
		Slug:  "project-001",
		Title: "Project 1",
	}
	db.Create(&project)
}

// init material model
func initMaterial(db *gorm.DB) {
	material := models.Material{
		Slug:       "material-001",
		Title:      "Material 1",
		Subtitle:   "Sub title Material 1",
		Min:        10,
		Max:        100,
		CategoryID: 1,
	}
	db.Create(&material)
}
