package main

import (
	"daijai/config"
	"daijai/models"
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	cleanFlag bool
	seedFlag  bool
)

func init() {
	config.ConnectDB()
	c := flag.Bool("clean", false, "Drop all tables")
	s := flag.Bool("seed", false, "Seed data")
	flag.Parse()
	cleanFlag = *c
	seedFlag = *s
}

// create flags
// -clean : drop all tables
// -seed : seed data

func main() {
	db := config.DB
	if dba, err := config.DB.DB(); err == nil {
		defer dba.Close()
	}
	tables := []interface{}{
		// relation tables
		&models.InventoryMaterial{},
		&models.InventoryMaterialTransaction{},
		&models.SumMaterialInventory{},
		&models.ReceiptMaterial{},
		&models.OrderBOM{},
		&models.OrderReserving{},
		&models.WithdrawalApprovement{},
		&models.WithdrawalTransaction{},
		&models.WithdrawalAdminTransaction{},
		&models.Withdrawal{},
		&models.PurchaseSuggestion{},
		&models.PurchaseMaterial{},

		// // main tables
		&models.Order{},
		&models.Material{},
		&models.AppLog{},
		&models.BOM{},
		&models.Category{},
		&models.Drawing{},
		&models.Inventory{},
		&models.Project{},
		&models.PurchasePORefs{},
		&models.Purchase{},
		&models.PORef{},
		&models.Slugger{},
		&models.Receipt{},
		&models.User{},
		&models.Notification{},
		&models.Adjustment{},
		&models.TransferMaterial{},
		&models.ProjectStore{},

		// extend tables
		&models.ExtendOrderBOM{},
		&models.ExtendOrder{},
		&models.ExtendOrderReserving{},
	}
	if cleanFlag {
		log.Println("Dropping all tables...")
		db.Migrator().DropTable()
		for _, table := range tables {
			db.Migrator().DropTable(table)
		}
		db.Migrator().DropTable(("building_areas"))
		db.Migrator().DropTable(("building_pets"))
		log.Println("All tables dropped")
	}

	log.Println("Migrating data...")
	for _, table := range tables {
		db.AutoMigrate(&table)
	}
	log.Println("Done! Migrating data ")

	if seedFlag {
		log.Println("Seeding data...")
		loadUsers(db, "./migrate/users.csv")
		initSlugger(db)
		initInventory(db)
		loadProjects(db, "./migrate/projects.csv")
		loadProjectStore(db, "./migrate/project_stores.csv")
		loadCategoriesFromCSV(db, "./migrate/categories.csv")

		// get all files in the materials folder
		files, err := ioutil.ReadDir("./migrate/materials")
		if err != nil {
			log.Println("Failed to read materials folder: ", err)
		}
		for _, file := range files {
			loadMaterialsFromCSV(db, fmt.Sprintf("./migrate/materials/%s", file.Name()))
			// loadMaterialsFromCSV(db, "./migrate/materials.csv")
		}
		// loadDrawingsFromCSV(db, "./migrate/drawings.csv")
		// loadMateriailOfDrawing(db, "./migrate/boms.csv")
		log.Println("Done! Seeding data")
	}
}

func initSlugger(db *gorm.DB) {
	slugables := []models.Slugable{
		&models.User{},
		&models.Order{},
		&models.Withdrawal{},
		&models.Purchase{},
		&models.Receipt{},
		&models.ExtendOrder{},
		&models.Drawing{},
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

func loadUsers(db *gorm.DB, filePath string) error {
	log.Println("Loading users from CSV file...")
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Failed to open CSV file: ", err)
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the CSV records
	records, err := reader.ReadAll()
	if err != nil {
		log.Println("Failed to read CSV records: ", err)
		return fmt.Errorf("failed to read CSV records: %w", err)
	}

	// Skip the first header row
	records = records[1:]

	// Process each record
	for _, record := range records {
		slug := record[0]
		username := record[1]
		pwd := record[2]
		fullName := record[3]
		role := record[4]
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err != nil {
			log.Println("Failed to hash password: ", err)
			continue
		}

		user := models.User{
			Slug:     slug,
			Username: username,
			Password: string(hashedPassword),
			FullName: fullName,
			Role:     role,
		}
		// Save the drawing to the database
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to save project to database: %w", err)
		}
	}
	log.Println("Users loaded successfully")
	return nil
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

func loadProjects(db *gorm.DB, filePath string) error {
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

	// Skip the first header row
	records = records[1:]

	// Process each record
	for _, record := range records {
		slug := strings.TrimSpace(record[0])
		title := record[1]
		subtitle := record[2]
		project := models.Project{
			Slug:  slug,
			Title: fmt.Sprintf("%s %s", title, subtitle),
		}
		// Save the drawing to the database
		if err := db.Create(&project).Error; err != nil {
			log.Printf("Failed to save project: %s error:  %v", slug, err)
			continue
		}
	}
	return nil
}

func loadProjectStore(db *gorm.DB, filePath string) error {
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

	// Skip the first header row
	records = records[1:]

	// Process each record
	for _, record := range records {
		slug := strings.TrimSpace(record[0])
		title := record[1]
		var project models.Project
		if err != nil {
			continue
		}

		if err := db.First(&project, "slug = ?", slug).Error; err != nil {
			log.Printf("Failed to find project: %s error:  %v", slug, err)
			continue
		}

		store := models.ProjectStore{
			Slug:      slug,
			Title:     title,
			ProjectID: project.ID,
		}
		// Save the drawing to the database
		if err := db.Create(&store).Error; err != nil {
			return fmt.Errorf("failed to save project to database: %w", err)
		}
	}
	return nil
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
			log.Println("failed to save material to database: %w", err)
			continue
		}
	}

	return nil
}

func loadMaterialsFromCSV(db *gorm.DB, filePath string) error {
	log.Println("Loading materials from CSV file: ", filePath)
	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		errMsg := fmt.Errorf("failed to open CSV file: %w", err)
		log.Println(errMsg.Error())
		return errMsg
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)
	// Read the CSV records
	records, err := reader.ReadAll()
	if err != nil {
		errMsg := fmt.Errorf("failed to read CSV records: %w", err)
		log.Println(errMsg.Error())
		return errMsg
	}

	// Skip the first header row
	records = records[1:]

	count := len(records)

	// Process each record
	log.Println("Total records: ", count)
	for _, record := range records {
		categorySlug := record[0]
		categoryTitle := record[1]
		slug := record[2]
		title := record[3]
		subtitle := record[4]
		supplier := record[5]
		defaultPrice, _ := strconv.ParseFloat(record[6], 64)
		isFG, _ := strconv.ParseBool(record[7])
		min, _ := strconv.Atoi(record[8])
		max, _ := strconv.Atoi(record[9])
		stock := 0
		rawStock := strings.TrimSpace(record[10])
		if rawStock != "" && rawStock != "-" {
			if s, err := strconv.Atoi(rawStock); err == nil {
				stock = s
			} else {
				log.Printf("%s - Failed to convert stock-value: %s ", slug, rawStock)
			}
		}

		if categorySlug == "" || slug == "" || title == "" {
			pV := fmt.Sprintf("categorySlug: %s, catTitle: %s, slug: %s, title: %s, subtitle: %s, supplier: %s, defaultPrice: %f, isFG: %t, min: %d, max: %d, stock: %d\n", categorySlug, categoryTitle, slug, title, subtitle, supplier, defaultPrice, isFG, min, max, stock)
			log.Println("🔥 Skip ⏭️", pV)
			continue
		}

		var categoryModel models.Category
		if err := db.Where(models.Category{Slug: categorySlug}).
			Attrs(models.Category{
				Slug:  categorySlug,
				Title: categoryTitle,
			}).
			FirstOrCreate(&categoryModel).
			Error; err != nil {
			errMsg := fmt.Errorf("failed to find category: %s  - id: %s", categorySlug, slug)
			log.Println(errMsg.Error())
			return errMsg
		}
		material := models.Material{
			CategoryID:   categoryModel.ID,
			Slug:         strings.TrimSpace(slug),
			Title:        strings.TrimSpace(title),
			Subtitle:     strings.TrimSpace(subtitle),
			Supplier:     strings.TrimSpace(supplier),
			DefaultPrice: int64(defaultPrice * 100),
			IsFG:         isFG,
			Min:          int64(min),
			Max:          int64(max),
			ImagePath:    fmt.Sprintf("/materials/%s.jpg", slug),
		}
		if err := db.Create(&material).Error; err != nil {
			log.Printf("🥶 Failed  %s: %v\n", slug, err)
			continue
		}

		// adjust stock
		outOfStock := stock == 0
		if outOfStock {
			continue
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			invt := uint(1)
			stock := int64(stock * 100)
			ppu := int64(defaultPrice * 100)
			adjustment := models.Adjustment{
				Quantity:     stock,
				MaterialID:   material.ID,
				InventoryID:  invt,
				CreatedByID:  1,
				PricePerUnit: ppu,
			}
			if err := tx.Create(&adjustment).Error; err != nil {
				log.Printf("Failed to create adjustment for material %s: %v\n", title, err)
				return nil
			}

			// create inventory material
			inventoryMaterial := models.InventoryMaterial{
				InventoryID:           invt,
				MaterialID:            material.ID,
				AdjustmentID:          &adjustment.ID,
				Quantity:              stock,
				AvailableQty:          stock,
				IsOutOfStock:          outOfStock,
				Price:                 ppu,
				InventoryMaterialType: models.InventoryMaterialType_Adjust,
			}
			if err := tx.Create(&inventoryMaterial).Error; err != nil {
				log.Printf("Failed to create inventory material for material %s: %v\n", title, err)
				return nil
			}
			// log.Printf("Material %s quantity: %d created \n", title, stock)

			var sumMaterialInventory models.SumMaterialInventory
			if err := db.
				Where("material_id = ?", material.ID).
				Where("inventory_id = ?", invt).
				FirstOrInit(&sumMaterialInventory).Error; err != nil {
				return err
			}

			sumMaterialInventory.MaterialID = material.ID
			sumMaterialInventory.InventoryID = invt
			sumMaterialInventory.Quantity = stock
			sumMaterialInventory.Price = ppu
			if err := db.Save(&sumMaterialInventory).Error; err != nil {
				log.Printf("Failed to save sum material inventory for material %s: %v\n", title, err)
				return nil
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func loadDrawingsFromCSV(db *gorm.DB, filePath string) error {
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

	// Skip the first header row
	records = records[1:]

	// Process each record
	for _, record := range records {
		slug := record[0]
		partNumber := record[1]
		imagePath := record[2]
		createdByID, _ := strconv.Atoi(record[3])
		isFG, _ := strconv.ParseBool(record[4])

		// Create a new drawing object
		drawing := models.Drawing{
			Slug:        slug,
			ImagePath:   imagePath,
			PartNumber:  partNumber,
			CreatedByID: uint(createdByID),
			IsFG:        isFG,
		}

		// Save the drawing to the database
		if err := db.Create(&drawing).Error; err != nil {
			return fmt.Errorf("failed to save drawing to database: %w", err)
		}
	}

	return nil
}

func loadMateriailOfDrawing(db *gorm.DB, filePath string) error {
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

	// Skip the first header row
	records = records[1:]

	// Process each record
	for _, record := range records {
		quantity, _ := strconv.Atoi(record[0])
		drawingID, _ := strconv.Atoi(record[1])
		materialID, _ := strconv.Atoi(record[2])

		// Create a new material of drawing object
		materialOfDrawing := models.BOM{
			DrawingID:  uint(drawingID),
			MaterialID: uint(materialID),
			Quantity:   int64(quantity),
		}

		// Save the material of drawing to the database
		if err := db.Create(&materialOfDrawing).Error; err != nil {
			return fmt.Errorf("failed to save material of drawing to database: %w", err)
		}
	}

	return nil
}
