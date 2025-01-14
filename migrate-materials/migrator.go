package main

import (
	"daijai/config"
	"daijai/models"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"gorm.io/gorm"
)

func init() {

}

// create flags
// -clean : drop all tables
// -seed : seed data

func main() {
	config.ConnectDB()
	db := config.DB
	if dba, err := config.DB.DB(); err == nil {
		defer dba.Close()
	}
	log.Println("Seeding data...")
	loadMaterialsFromCSV(db, "./migrate-materials/materials.csv")
	log.Println("Done! Seeding data")
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

	// Skip the first header row
	records = records[1:]

	// Process each record
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
		stock, _ := strconv.Atoi(record[10])
		fmt.Printf("categorySlug: %s, catTitle: %s, slug: %s, title: %s, subtitle: %s, supplier: %s, defaultPrice: %f, isFG: %t, min: %d, max: %d, stock: %d\n", categorySlug, categoryTitle, slug, title, subtitle, supplier, defaultPrice, isFG, min, max, stock)

		if categorySlug == "" || slug == "" || title == "" {
			log.Println("Skipping record due to missing required fields")
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
			return fmt.Errorf("failed to find or init category: %w", err)
		}
		material := models.Material{
			CategoryID:   categoryModel.ID,
			Slug:         slug,
			Title:        title,
			Subtitle:     subtitle,
			Supplier:     supplier,
			DefaultPrice: int64(defaultPrice * 100),
			IsFG:         isFG,
			Min:          int64(min),
			Max:          int64(max),
			ImagePath:    fmt.Sprintf("/materials/%s.jpg", slug),
		}
		if err := db.Create(&material).Error; err != nil {
			log.Printf("Failed to create material %s: %v\n", title, err)
			continue
		}

		// adjust stock

		outOfStock := stock == 0
		if outOfStock {
			log.Printf("Material %s is out of stock \n", title)
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
			log.Printf("Material %s quantity: %d created \n", title, stock)

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
