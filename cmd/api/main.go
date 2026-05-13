package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/muktiabdii/pdf-management-api/internal/router"
	"github.com/muktiabdii/pdf-management-api/pkg/database"
	"github.com/muktiabdii/pdf-management-api/pkg/storage"
)

func main() {
	// 1. load .env
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	// 2. init database
	database.Connect()
	database.Migrate()

	// 3. init S3
	storage.Connect()

	// 4. setup router
	r := router.Setup(database.DB)

	// 5. jalankan server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}