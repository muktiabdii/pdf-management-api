package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/muktiabdii/pdf-management-api/internal/router"
	"github.com/muktiabdii/pdf-management-api/pkg/database"
	"github.com/muktiabdii/pdf-management-api/pkg/storage"
)

// main initializes and starts the PDF Management API server.
// It performs the following initialization steps:
//  1. Loads environment variables from .env file
//  2. Establishes database connection and runs migrations
//  3. Initializes AWS S3 storage connection
//  4. Sets up HTTP routes and dependency injection
//  5. Starts the HTTP server on the configured port (default: 8080)
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	database.Connect()
	database.Migrate()

	storage.Connect()

	r := router.Setup(database.DB)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server running on port %s", port)
	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
