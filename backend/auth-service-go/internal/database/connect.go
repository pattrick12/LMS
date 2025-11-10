package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDatabase() (*gorm.DB, error) {
	// --- Hardcoded connection details for the Docker development environment ---
	// The host must be the name of the Docker container.
	host := "lms-postgres"
	user := "postgres"
	password := "postgres" // This is set in your docker run command
	dbname := "postgres"   // Default database name
	port := "5432"         // Default postgres port

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		host, user, password, dbname, port,
	)

	// Configure the logger to be less verbose
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel: logger.Warn, // Only log warnings and errors
			Colorful: true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}

	log.Println("✅ Successfully connected to PostgreSQL database with GORM!")
	return db, nil
}
