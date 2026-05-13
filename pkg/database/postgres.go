package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/muktiabdii/pdf-management-api/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database connection instance.
var DB *gorm.DB

// Connect establishes a connection to the PostgreSQL database.
// It implements retry logic (5 attempts with 2-second intervals) to handle
// database startup delays in containerized environments.
func Connect() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	var db *gorm.DB
	var err error

	// Attempt to connect with retry logic
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			break
		}
		log.Printf("database not ready, retrying in 2 seconds... (%d/5)", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("failed to connect to database after 5 attempts: %v", err)
	}

	DB = db
	log.Println("database connected successfully")
}

// Migrate runs all pending database migrations.
// It should be called after establishing a database connection to ensure
// all required tables and schemas are created.
func Migrate() {
	if DB == nil {
		log.Fatal("database not connected, call Connect() first")
	}

	err := DB.AutoMigrate(&model.PdfFile{})
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("migration completed successfully")
}
