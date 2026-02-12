package database

import (
	"atoyr/server/internal/models"
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	migratedModels = []any{
		&models.SessionEntity{},
		&models.UserEntity{},
		&models.InventoryEntity{},
		&models.InventoryItemEntity{},
		&models.ResultEntity{},
		&models.UserResultEntity{},
	}
)

func Initialize() (*gorm.DB, error) {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		return nil, fmt.Errorf("DATABASE_PATH environment variable is not set")
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Determine log level based on environment
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate models
	if err := Automigrate(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func Automigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(migratedModels...); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}
