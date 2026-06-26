package database

import (
	"log"

	"github.com/Sarthak-Nagaria/ticket-system/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Init opens the SQLite database at dbPath, runs auto-migrations for all
// models, and returns the *gorm.DB handle to be used by the application.
func Init(dbPath string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Ticket{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	log.Println("Database connected and migrated successfully")
	return db
}
