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
	db := config.DB
	migrator := db.Migrator()
	tables := []interface{}{
		// relation tables
		&models.ReceiptMaterial{},
		&models.OrderBom{},
		&models.OrderReserving{},
		&models.WithdrawalTransaction{},
		&models.InventoryMaterial{},
		&models.InventoryMaterialTransaction{},
		&models.PurchaseSuggestion{},
		&models.PurchaseMaterial{},
		// main tables
		&models.Withdrawal{},
		&models.Receipt{},
		&models.Order{},
		&models.Material{},
		&models.AppLog{},
		&models.Bom{},
		&models.Category{},
		&models.Drawing{},
		&models.Inventory{},
		&models.Project{},
		&models.PurchasePORefs{},
		&models.Purchase{},
		&models.PORef{},
		&models.Slugger{},
		&models.User{},
		&models.Notification{},
	}
	for _, table := range tables {
		migrator.DropTable(table)
		db.AutoMigrate(&table)
	}

	initUsers(config.DB)
	initCategory(config.DB)
	initInventory(config.DB)
	initProject(config.DB)
	initMaterial(config.DB)
	initSlugger(config.DB)
}

func initSlugger(db *gorm.DB) {
	slugables := []models.Slugable{
		&models.User{},
		&models.Order{},
		&models.Withdrawal{},
		&models.Purchase{},
	}
	for _, m := range slugables {
		slug := m.GenerateSlug()
		s := models.Slugger{
			TableName: slug.TableName,
			Prefix:    slug.Prefix,
			Pad:       slug.Pad,
			Value:     0,
		}
		db.Create(&s)
	}
}

func initUsers(db *gorm.DB) {
	pwd := "$2a$10$zeswe0/DbG/2k.4KlIbLTO2bYwmvpbXMYp2aJf.dyy7FXyHOmg9xm"
	users := []models.User{
		{
			Slug:     "ADM-001",
			Username: "salah",
			Password: pwd,
			FullName: "John Doe",
			Role:     models.ROLE_Admin,
			Tel:      "0990938983",
		},
		{
			Slug:     "TCH-001",
			Username: "woofoo",
			Password: pwd,
			FullName: "Woo Foo",
			Role:     models.ROLE_Tech,
			Tel:      "0994441111",
		},
	}

	for _, user := range users {
		db.Create(&user)
	}
}

func initCategory(db *gorm.DB) {
	categories := []models.Category{
		{
			Slug:     "ไม่้บอร์ด",
			Title:    "Furniture Material 1",
			Subtitle: "Mock Furniture Material 1",
		},
		{
			Slug:     "ปิดขอบ",
			Title:    "Furniture Material 2",
			Subtitle: "Mock Furniture Material 2",
		},
		{
			Slug:     "น็อต สรู",
			Title:    "Furniture Material 3",
			Subtitle: "Mock Furniture Material 3",
		},
		{
			Slug:     "มือจับ",
			Title:    "Furniture Material 4",
			Subtitle: "Mock Furniture Material 4",
		},
		{
			Slug:     "ลิ้นชัก",
			Title:    "Furniture Material 5",
			Subtitle: "Mock Furniture Material 5",
		},
	}

	for _, category := range categories {
		db.Create(&category)
	}
}

func initInventory(db *gorm.DB) {
	inventories := []models.Inventory{
		{
			Slug:  "IVT-001",
			Title: "Main Inventory",
		},
		{
			Slug:  "IVT-002",
			Title: "Factory Inventory",
		},
	}

	for _, inventory := range inventories {
		db.Create(&inventory)
	}
}

// init project model
func initProject(db *gorm.DB) {
	projects := []models.Project{
		{
			Slug:  "PRJ-001",
			Title: "NAWAMIN",
		},
		{
			Slug:  "PRJ-002",
			Title: "RamIndhra",
		},
	}

	for _, v := range projects {
		db.Create(&v)
	}
}

// init material model
func initMaterial(db *gorm.DB) {
	materials := []models.Material{
		{
			Slug:       "MAT-001",
			Title:      "ไม้อัด 20mm",
			Subtitle:   "ไม้ยางพาราอัด หน้าขาว",
			Min:        10,
			Max:        100,
			CategoryID: 1,
		},
		{
			Slug:       "MAT-002",
			Title:      "ฟอเมก้า ปิดขอบ 2mm.",
			Subtitle:   "สีไม้ ความยาว 20เมตร",
			Min:        10,
			Max:        100,
			CategoryID: 1,
		},
	}
	for _, mat := range materials {
		db.Create(&mat)
	}
}
