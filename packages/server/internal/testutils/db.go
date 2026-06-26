package testutils

import (
	"atoyr/server/internal/database"
	"fmt"
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

	if err := database.Automigrate(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}
