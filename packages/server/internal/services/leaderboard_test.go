package services

import (
	"testing"

	"atoyr/server/internal/models"
	"atoyr/server/internal/testutils"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestLeaderboardService_GetLeaderboard(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewLeaderboardService(db)

	// Add some test results
	results := []models.SessionEntity{
		{
			ID:                       "session1",
			Score:                    100,
			TotalAttempts:            10,
			Accuracy:                 90,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    SessionPhaseFinished,
		},
		{
			ID:                       "session2",
			Score:                    150,
			TotalAttempts:            15,
			Accuracy:                 95,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    SessionPhaseFinished,
		},
		{
			ID:                       "session3",
			Score:                    80,
			TotalAttempts:            8,
			Accuracy:                 85,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    SessionPhaseFinished,
		},
		{
			ID:                       "session4",
			Score:                    0,
			TotalAttempts:            8,
			Accuracy:                 0,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    SessionPhaseFinished,
		},
	}

	for _, result := range results {
		db.Create(&result)
	}

	// Get leaderboard
	entries, err := service.GetLeaderboard(10, 0)
	assert.NoError(t, err)
	assert.Len(t, entries, 3)

	// Verify sorting by score (descending)
	assert.Equal(t, int32(150), entries[0].Score)
	assert.Equal(t, int32(100), entries[1].Score)
	assert.Equal(t, int32(80), entries[2].Score)
}

func TestLeaderboardService_GetLeaderboard_SameScore(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewLeaderboardService(db)

	// Add some test results
	results := []models.SessionEntity{
		{
			ID:                       "s01",
			Score:                    100,
			TotalAttempts:            10,
			Accuracy:                 90,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    SessionPhaseFinished,
		},
		{
			ID:                       "s02",
			Score:                    100,
			TotalAttempts:            9,
			Accuracy:                 95,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    SessionPhaseFinished,
		},
	}

	for _, result := range results {
		db.Create(&result)
	}

	// Get leaderboard
	entries, err := service.GetLeaderboard(10, 0)
	assert.NoError(t, err)
	assert.Len(t, entries, 2)

	// Verify sorting by score, then accuracy (descending)
	assert.Equal(t, "s02", entries[0].ID)
	assert.Equal(t, float32(95), entries[0].Accuracy)

	assert.Equal(t, "s01", entries[1].ID)
	assert.Equal(t, float32(90), entries[1].Accuracy)
}

func TestLeaderboardService_Pagination(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewLeaderboardService(db)

	// Add 25 results
	for i := 0; i < 25; i++ {
		result := models.SessionEntity{
			ID:                       "session" + string(rune(i)),
			Score:                    int32(100 + i),
			TotalAttempts:            int32(10),
			Accuracy:                 90,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    SessionPhaseFinished,
		}
		db.Create(&result)
	}

	// Get first page (limit 10)
	entries, _ := service.GetLeaderboard(10, 0)
	assert.Len(t, entries, 10)

	// Get second page
	entries, _ = service.GetLeaderboard(10, 10)
	assert.Len(t, entries, 10)

	// Get third page (should have 5)
	entries, _ = service.GetLeaderboard(10, 20)
	assert.Len(t, entries, 5)
}

func TestLeaderboardService_GetTotalEntries(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewLeaderboardService(db)

	total, _ := service.GetTotalEntries()
	assert.Equal(t, int64(0), total)

	// Add 5 results
	for i := 0; i < 5; i++ {
		result := models.SessionEntity{
			ID:                       "session" + string(rune(i)),
			Score:                    int32(100 + i),
			TotalAttempts:            10,
			Accuracy:                 90,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    SessionPhaseFinished,
		}
		db.Create(&result)
	}

	total, _ = service.GetTotalEntries()
	assert.Equal(t, int64(5), total)
}

func TestLeaderboardService_GetTopScores(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewLeaderboardService(db)

	// Add results
	scores := []int32{50, 100, 150, 75, 200}
	for i, score := range scores {
		result := models.SessionEntity{
			ID:                       "session" + string(rune(i)),
			Score:                    score,
			TotalAttempts:            10,
			Accuracy:                 90,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    SessionPhaseFinished,
		}
		db.Create(&result)
	}

	// Get top 3
	entries, _ := service.GetTopScores(3)
	assert.Len(t, entries, 3)

	// Verify top scores
	assert.Equal(t, int32(200), entries[0].Score)
	assert.Equal(t, int32(150), entries[1].Score)
	assert.Equal(t, int32(100), entries[2].Score)
}
