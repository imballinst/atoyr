package services

import (
	"testing"
	"time"

	"atoyr/server/internal/models"
)

func TestLeaderboardService_GetLeaderboard(t *testing.T) {
	db := setupTestDB(t)
	service := NewLeaderboardService(db)

	// Add some test results
	results := []models.ResultEntity{
		{
			SessionID:     "session1",
			Score:         100,
			TotalAttempts: 10,
			Accuracy:      90,
			Timestamp:     time.Now(),
		},
		{
			SessionID:     "session2",
			Score:         150,
			TotalAttempts: 15,
			Accuracy:      95,
			Timestamp:     time.Now(),
		},
		{
			SessionID:     "session3",
			Score:         80,
			TotalAttempts: 8,
			Accuracy:      85,
			Timestamp:     time.Now(),
		},
	}

	for _, result := range results {
		db.Create(&result)
	}

	// Get leaderboard
	entries, err := service.GetLeaderboard(10, 0)
	if err != nil {
		t.Fatalf("Failed to get leaderboard: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Verify sorting by score (descending)
	if entries[0].Score != 150 {
		t.Errorf("Expected highest score 150, got %d", entries[0].Score)
	}

	if entries[1].Score != 100 {
		t.Errorf("Expected second score 100, got %d", entries[1].Score)
	}

	if entries[2].Score != 80 {
		t.Errorf("Expected lowest score 80, got %d", entries[2].Score)
	}
}

func TestLeaderboardService_Pagination(t *testing.T) {
	db := setupTestDB(t)
	service := NewLeaderboardService(db)

	// Add 25 results
	for i := 0; i < 25; i++ {
		result := models.ResultEntity{
			SessionID:     "session" + string(rune(i)),
			Score:         int32(100 + i),
			TotalAttempts: int32(10),
			Accuracy:      90,
			Timestamp:     time.Now(),
		}
		db.Create(&result)
	}

	// Get first page (limit 10)
	entries, _ := service.GetLeaderboard(10, 0)
	if len(entries) != 10 {
		t.Errorf("Expected 10 entries in first page, got %d", len(entries))
	}

	// Get second page
	entries, _ = service.GetLeaderboard(10, 10)
	if len(entries) != 10 {
		t.Errorf("Expected 10 entries in second page, got %d", len(entries))
	}

	// Get third page (should have 5)
	entries, _ = service.GetLeaderboard(10, 20)
	if len(entries) != 5 {
		t.Errorf("Expected 5 entries in third page, got %d", len(entries))
	}
}

func TestLeaderboardService_GetTotalEntries(t *testing.T) {
	db := setupTestDB(t)
	service := NewLeaderboardService(db)

	total, _ := service.GetTotalEntries()
	if total != 0 {
		t.Errorf("Expected 0 entries initially, got %d", total)
	}

	// Add 5 results
	for i := 0; i < 5; i++ {
		result := models.ResultEntity{
			SessionID:     "session" + string(rune(i)),
			Score:         int32(100 + i),
			TotalAttempts: 10,
			Accuracy:      90,
			Timestamp:     time.Now(),
		}
		db.Create(&result)
	}

	total, _ = service.GetTotalEntries()
	if total != 5 {
		t.Errorf("Expected 5 entries, got %d", total)
	}
}

func TestLeaderboardService_GetTopScores(t *testing.T) {
	db := setupTestDB(t)
	service := NewLeaderboardService(db)

	// Add results
	scores := []int32{50, 100, 150, 75, 200}
	for i, score := range scores {
		result := models.ResultEntity{
			SessionID:     "session" + string(rune(i)),
			Score:         score,
			TotalAttempts: 10,
			Accuracy:      90,
			Timestamp:     time.Now(),
		}
		db.Create(&result)
	}

	// Get top 3
	entries, _ := service.GetTopScores(3)
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Verify top scores
	if entries[0].Score != 200 {
		t.Errorf("Expected top score 200, got %d", entries[0].Score)
	}

	if entries[1].Score != 150 {
		t.Errorf("Expected second score 150, got %d", entries[1].Score)
	}

	if entries[2].Score != 100 {
		t.Errorf("Expected third score 100, got %d", entries[2].Score)
	}
}
