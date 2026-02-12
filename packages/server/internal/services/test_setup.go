package services

import (
	"testing"

	"atoyr/server/internal/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := database.Automigrate(db); err != nil {
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
