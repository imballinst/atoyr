package services

import (
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
		words: []WordDefinition{
			{Word: "hello"},
			{Word: "world"},
			{Word: "apple"},
			{Word: "banana"},
			{Word: "cherry"},
			{Word: "dragon"},
			{Word: "elephant"},
			{Word: "forest"},
			{Word: "guitar"},
			{Word: "horizon"},
		},
	}
	return ws
}
