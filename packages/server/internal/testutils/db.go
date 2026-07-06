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

	// Find the migrations directory relative to this file
	_, currentFile, _, _ := runtime.Caller(0)
	serverDir := filepath.Join(filepath.Dir(currentFile), "..", "..")
	migrationsDir := filepath.Join(serverDir, "migrations")

	// Use migration runner for tests as well
	if err := database.RunMigrationsWithPath(db, migrationsDir); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}
