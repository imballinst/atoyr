package services

import (
	"testing"
	"time"

	"atoyr/server/internal/testutils"
	"atoyr/server/internal/utils"

	"github.com/stretchr/testify/assert"
)

func TestLeaderboardService_GetLeaderboard(t *testing.T) {
	sessionService := NewSessionService(testutils.SetupTestDB(t))
	service := NewLeaderboardService(sessionService, &FakeLeaderboardClient{
		Entries: []LeaderboardEntry{
			{ID: "s2", Rank: 1, Score: 150, TotalAttempts: 15, Accuracy: 95, Timestamp: time.Now().UnixMilli()},
			{ID: "s1", Rank: 2, Score: 100, TotalAttempts: 10, Accuracy: 90, Timestamp: time.Now().UnixMilli()},
			{ID: "s3", Rank: 3, Score: 80, TotalAttempts: 8, Accuracy: 85.12, Timestamp: time.Now().UnixMilli()},
		},
		Total: 3,
	})

	entries, err := service.GetLeaderboard("vanilla", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, entries, 3)
	assert.Equal(t, int32(150), entries[0].Score)
	assert.Equal(t, int32(100), entries[1].Score)
	assert.Equal(t, int32(80), entries[2].Score)
}

func TestLeaderboardService_GetLeaderboard_Pagination(t *testing.T) {
	entries := make([]LeaderboardEntry, 25)
	for i := range 25 {
		entries[i] = LeaderboardEntry{
			ID:    "s" + utils.GenerateUUID(),
			Rank:  int32(i + 1),
			Score: int32(100 + i),
		}
	}

	sessionService := NewSessionService(testutils.SetupTestDB(t))
	service := NewLeaderboardService(sessionService, &FakeLeaderboardClient{
		Entries: entries,
		Total:   25,
	})

	page1, _ := service.GetLeaderboard("vanilla", 10, 0)
	assert.Len(t, page1, 10)

	page2, _ := service.GetLeaderboard("vanilla", 10, 10)
	assert.Len(t, page2, 10)

	page3, _ := service.GetLeaderboard("vanilla", 10, 20)
	assert.Len(t, page3, 5)
}

func TestLeaderboardService_GetTotalEntries(t *testing.T) {
	sessionService := NewSessionService(testutils.SetupTestDB(t))
	service := NewLeaderboardService(sessionService, &FakeLeaderboardClient{
		Total: 5,
	})

	total, err := service.GetTotalEntries("vanilla")
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
}

func TestLeaderboardService_GetPercentile(t *testing.T) {
	db := testutils.SetupTestDB(t)
	sessionService := NewSessionService(db)
	session, err := sessionService.Create(false, []string{}, "vanilla", 30)
	assert.NoError(t, err)
	db.Exec("UPDATE session_entities SET user_id = ? WHERE id = ?", "user-a", session.ID)

	service := NewLeaderboardService(sessionService, &FakeLeaderboardClient{
		Percentile: 42.5,
	})

	percentile, err := service.GetPercentile(session.ID, "vanilla")
	assert.NoError(t, err)
	assert.InDelta(t, float32(42.5), percentile, 0.001)
}

func TestLeaderboardService_GetPercentile_NoUserID(t *testing.T) {
	db := testutils.SetupTestDB(t)
	sessionService := NewSessionService(db)
	session, err := sessionService.Create(false, []string{}, "vanilla", 30)
	assert.NoError(t, err)

	service := NewLeaderboardService(sessionService, &FakeLeaderboardClient{})

	_, err = service.GetPercentile(session.ID, "vanilla")
	assert.ErrorContains(t, err, "no user id")
}

func TestLeaderboardService_UpsertEntry(t *testing.T) {
	db := testutils.SetupTestDB(t)
	sessionService := NewSessionService(db)
	session, err := sessionService.Create(false, []string{}, "vanilla", 30)
	assert.NoError(t, err)
	db.Exec("UPDATE session_entities SET user_id = ? WHERE id = ?", "user-a", session.ID)

	service := NewLeaderboardService(sessionService, &FakeLeaderboardClient{})

	err = service.UpsertEntry(session)
	assert.NoError(t, err)
}

func TestLeaderboardService_UpsertEntry_NoUserID(t *testing.T) {
	db := testutils.SetupTestDB(t)
	sessionService := NewSessionService(db)
	session, err := sessionService.Create(false, []string{}, "vanilla", 30)
	assert.NoError(t, err)

	service := NewLeaderboardService(sessionService, &FakeLeaderboardClient{})

	err = service.UpsertEntry(session)
	assert.NoError(t, err)
}
