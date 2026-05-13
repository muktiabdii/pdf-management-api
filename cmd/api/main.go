package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/muktiabdii/pdf-management-api/pkg/database"
	"github.com/muktiabdii/pdf-management-api/pkg/storage"
)

func main() {
	// load .env
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	// init database
	database.Connect()
	database.Migrate()

	// init S3
	storage.Connect()

	log.Println("application started successfully")
}