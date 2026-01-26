package services

import (
	"os"
	"testing"

	"atoyr/server/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.AutoMigrate(&models.SessionEntity{}, &models.ResultEntity{}); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

func setupTestWordService(t *testing.T) *WordService {
	ws := &WordService{
		words: []string{
			"hello", "world", "apple", "banana", "cherry",
			"dragon", "elephant", "forest", "guitar", "horizon",
		},
	}
	return ws
}
