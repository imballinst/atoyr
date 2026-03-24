package testutils

import (
	"atoyr/server/internal/database"
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	gormConfig := gorm.Config{}

	if os.Getenv("GORM_DEBUG") == "true" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gormConfig)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := database.Automigrate(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}
