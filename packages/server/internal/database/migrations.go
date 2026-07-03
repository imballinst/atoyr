package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" && os.Getenv("ENV") == "production" {
		panic("MIGRATIONS_DIR should be provided in production mode")
	} else {
		migrationsDir = "migrations"
	}

	return RunMigrationsWithPath(db, migrationsDir)
}

func RunMigrationsWithPath(db *gorm.DB, migrationsDir string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	return runMigrationsWithSQLDB(sqlDB, migrationsDir)
}

func runMigrationsWithSQLDB(sqlDB *sql.DB, migrationsDir string) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Create file source
	sourceDriver := &file.File{}
	source, err := sourceDriver.Open(fmt.Sprintf("file://%s", absPath))
	if err != nil {
		return fmt.Errorf("failed to open migration source: %w", err)
	}

	// Create sqlite3 database driver
	dbDriver, err := sqlite3.WithInstance(sqlDB, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("file", source, "sqlite3", dbDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
