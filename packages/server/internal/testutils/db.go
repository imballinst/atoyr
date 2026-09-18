package testutils

import (
	"atoyr/server/internal/database"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupTestDB(t testing.TB) *gorm.DB {
	return SetupTestDBWithVersion(t, 0)
}

func SetupTestDBWithVersion(t testing.TB, version uint) *gorm.DB {
	gormConfig := gorm.Config{}

	if os.Getenv("GORM_DEBUG") == "true" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	dsn := fmt.Sprintf("file:test_%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gormConfig)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	sqlDb, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get database: %v", err)
	}
	sqlDb.SetMaxOpenConns(1)

	// Use migration runner for tests as well
	if err := RunMigrations(t, db, version); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

func RunMigrations(t testing.TB, db *gorm.DB, version uint) error {
	// Find the migrations directory relative to this file
	_, currentFile, _, _ := runtime.Caller(0)
	serverDir := filepath.Join(filepath.Dir(currentFile), "..", "..")
	migrationsDir := filepath.Join(serverDir, "migrations")

	// Use migration runner for tests as well
	return database.RunMigrationsWithPath(db, migrationsDir, version)
}
