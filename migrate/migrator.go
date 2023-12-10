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
	// config.DB.Migrator().DropTable(&models.Project{})
	// config.DB.Migrator().DropTable(&models.Category{})
	// config.DB.Migrator().DropTable(&models.Material{})
	// config.DB.Migrator().DropTable(&models.Drawing{})
	// config.DB.Migrator().DropTable(&models.Bomb{})
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
	// config.DB.AutoMigrate(&models.Drawing{})
	// config.DB.AutoMigrate(&models.Bomb{})
	config.DB.AutoMigrate(&models.Receipt{})
	config.DB.AutoMigrate(&models.ReceiptMaterial{})
	// config.DB.AutoMigrate(&models.Withdrawal{})
	// config.DB.AutoMigrate(&models.WithdrawalMaterial{})
	// config.DB.AutoMigrate(&models.Purchase{})
	// config.DB.AutoMigrate(&models.PurchaseMaterial{})

	// initAdmin(config.DB)
	// initUser(config.DB)
	// initCategory(config.DB)
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
