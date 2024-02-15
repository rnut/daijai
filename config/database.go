package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	// Retrieve database credentials from environment variables
	// host := os.Getenv("DB_HOST")
	// port := os.Getenv("DB_PORT")
	// user := os.Getenv("DB_USER")
	// password := os.Getenv("DB_PASSWORD")
	// dbName := os.Getenv("DB_NAME")
	// sslMode := os.Getenv("DB_SSL_MODE")
	// timezone := os.Getenv("DB_TIMEZONE")

	host := "ep-spring-rain-18759799.ap-southeast-1.aws.neon.tech"
	port := "5432"
	user := "arnut.khu"
	password := "4QA9LDPZhjHF"
	dbName := "daijai"
	sslMode := "require"
	timezone := "Asia/Bangkok"

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		host, port, user, password, dbName, sslMode, timezone)

	log.Println(dsn)
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Println("ERROR: cannot connect to the database")
	} else {
		log.Println("Connected to the database successfully!")
	}
}

func GetDB() *gorm.DB {
	return DB
}
