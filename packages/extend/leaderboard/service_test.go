package leaderboard

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedStore() *InMemoryStore {
	store := NewInMemoryStore()
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	_ = store.AddSession("session-1", "user-a", "vanilla", base)
	_ = store.AddSession("session-2", "user-b", "vanilla", base.Add(time.Hour))
	_ = store.UpsertEntry("vanilla", Entry{UserID: "user-a", Score: 100, TotalAttempts: 10, Accuracy: 0.8, FinishedAt: base.Add(2 * time.Hour)})
	_ = store.UpsertEntry("vanilla", Entry{UserID: "user-b", Score: 90, TotalAttempts: 12, Accuracy: 0.75, FinishedAt: base.Add(3 * time.Hour)})
	_ = store.UpsertEntry("vanilla", Entry{UserID: "user-c", Score: 100, TotalAttempts: 8, Accuracy: 0.9, FinishedAt: base.Add(time.Hour)})
	return store
}

func TestInMemoryStore(t *testing.T) {
	t.Run("leaderboard ranks by score, accuracy, and finished-at", func(t *testing.T) {
		store := seedStore()
		entries, total, err := store.ListEntries("vanilla", 10, 0)
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		require.Len(t, entries, 3)
		assert.Equal(t, "user-c", entries[0].UserID)
		assert.Equal(t, "user-a", entries[1].UserID)
		assert.Equal(t, "user-b", entries[2].UserID)
	})

	t.Run("percentile excludes the current user", func(t *testing.T) {
		store := seedStore()
		percentile, err := store.Percentile("vanilla", "user-c")
		require.NoError(t, err)
		assert.InDelta(t, 100.0, percentile, 0.001)

		percentile, err = store.Percentile("vanilla", "user-b")
		require.NoError(t, err)
		assert.InDelta(t, 0.0, percentile, 0.001)

		percentile, err = store.Percentile("vanilla", "user-a")
		require.NoError(t, err)
		assert.InDelta(t, 50.0, percentile, 0.001)
	})

	t.Run("active-session registry tracks add and remove", func(t *testing.T) {
		store := seedStore()
		sessions := store.ListActiveSessions()
		assert.Len(t, sessions, 2)

		_ = store.RemoveSession("session-1")
		sessions = store.ListActiveSessions()
		require.Len(t, sessions, 1)
		assert.Equal(t, "session-2", sessions[0].ID)
	})
}

func TestServer_GetLeaderboard(t *testing.T) {
	server := httptest.NewServer(NewServer(seedStore()).Handler())
	defer server.Close()

	resp, err := http.Get(server.URL + "/leaderboard?mode=vanilla&limit=2&offset=0")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body LeaderboardResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, 3, body.Total)
	assert.Len(t, body.Entries, 2)
	assert.Equal(t, 1, body.Entries[0].Rank)
	assert.Equal(t, "user-c", body.Entries[0].UserID)
	assert.Equal(t, 2, body.Entries[1].Rank)
	assert.Equal(t, "user-a", body.Entries[1].UserID)
}

func TestServer_GetPercentile(t *testing.T) {
	server := httptest.NewServer(NewServer(seedStore()).Handler())
	defer server.Close()

	resp, err := http.Get(server.URL + "/leaderboard/percentile?mode=vanilla&userId=user-a")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body PercentileResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.InDelta(t, 50.0, body.Percentile, 0.001)
}

func TestServer_GetPercentile_UserNotFound(t *testing.T) {
	server := httptest.NewServer(NewServer(seedStore()).Handler())
	defer server.Close()

	resp, err := http.Get(server.URL + "/leaderboard/percentile?mode=vanilla&userId=user-z")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestServer_GetActiveSessions(t *testing.T) {
	server := httptest.NewServer(NewServer(seedStore()).Handler())
	defer server.Close()

	resp, err := http.Get(server.URL + "/active-sessions")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body ActiveSessionsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Len(t, body.Sessions, 2)
}

func TestServer_GetLeaderboard_MissingMode(t *testing.T) {
	server := httptest.NewServer(NewServer(NewInMemoryStore()).Handler())
	defer server.Close()

	resp, err := http.Get(server.URL + "/leaderboard?limit=10")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestServer_GetPercentile_MissingParams(t *testing.T) {
	server := httptest.NewServer(NewServer(NewInMemoryStore()).Handler())
	defer server.Close()

	resp, err := http.Get(server.URL + "/leaderboard/percentile")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
