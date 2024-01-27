package main

import (
	"daijai/config"
	"daijai/models"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

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
	loadCategoriesFromCSV(config.DB, "./migrate/categories.csv")
	loadMaterialsFromCSV(config.DB, "./migrate/materials.csv")
	initInventory(config.DB)
	initProject(config.DB)
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

func loadCategoriesFromCSV(db *gorm.DB, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the CSV records
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV records: %w", err)
	}

	// Process each record
	for _, record := range records {
		slug := record[0]
		title := record[1]
		subtitle := record[2]
		isFG, _ := strconv.ParseBool(record[3])

		category := models.Category{
			Slug:     slug,
			Title:    title,
			Subtitle: subtitle,
			IsFG:     isFG,
		}

		if err := db.Create(&category).Error; err != nil {
			return fmt.Errorf("failed to save material to database: %w", err)
		}
	}

	return nil
}

func loadMaterialsFromCSV(db *gorm.DB, filePath string) error {
	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the CSV records
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV records: %w", err)
	}

	// Process each record
	for _, record := range records {
		slug := record[0]
		title := record[1]
		subtitle := record[2]
		categoryID, _ := strconv.Atoi(record[3])
		isFG, _ := strconv.ParseBool(record[4])

		// Create a new material object
		material := models.Material{
			Slug:       slug,
			Title:      title,
			Subtitle:   subtitle,
			Min:        0,
			Max:        0,
			CategoryID: uint(categoryID),
			IsFG:       isFG,
		}

		// Save the material to the database
		if err := db.Create(&material).Error; err != nil {
			return fmt.Errorf("failed to save material to database: %w", err)
		}
	}

	return nil
}
