package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectDatabase sets up the connection to the PostgreSQL database using GORM.
func ConnectDatabase() (*gorm.DB, error) {
	// Hardcoded connection details for the Docker development environment
	host := "lms-postgres"
	user := "postgres"
	password := "postgres"
	dbname := "postgres"
	port := "5432"

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		host, user, password, dbname, port,
	)

	// Configure GORM logger for cleaner output (only warnings and errors)
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel: logger.Warn,
			Colorful: true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}

	log.Println("✅ [ERP Service] Successfully connected to PostgreSQL with GORM!")
	return db, nil
}
