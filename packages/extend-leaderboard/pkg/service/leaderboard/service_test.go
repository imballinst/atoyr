package leaderboard

import (
	"context"
	"net"
	"sort"
	"sync"
	"testing"
	"time"

	leaderboardpb "extend-leaderboard/pkg/pb/leaderboard/v1"
	"extend-leaderboard/pkg/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type inMemoryStore struct {
	mu       sync.RWMutex
	entries  map[string]map[string]storage.Entry
	sessions map[string]storage.ActiveSession
}

func newInMemoryStore() *inMemoryStore {
	return &inMemoryStore{
		entries:  make(map[string]map[string]storage.Entry),
		sessions: make(map[string]storage.ActiveSession),
	}
}

func (s *inMemoryStore) AddSession(_ context.Context, id, userID, mode string, startedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = storage.ActiveSession{ID: id, UserID: userID, Mode: mode, StartedAt: startedAt}
	return nil
}

func (s *inMemoryStore) RemoveSession(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
	return nil
}

func (s *inMemoryStore) ListActiveSessions(_ context.Context) ([]storage.ActiveSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sessions := make([]storage.ActiveSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (s *inMemoryStore) UpsertEntry(_ context.Context, mode string, entry storage.Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.entries[mode] == nil {
		s.entries[mode] = make(map[string]storage.Entry)
	}
	s.entries[mode][entry.UserID] = entry
	return nil
}

func (s *inMemoryStore) ListEntries(_ context.Context, mode string, limit, offset int) ([]storage.Entry, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := s.sortedEntries(mode)
	total := len(all)
	if offset > total {
		return []storage.Entry{}, total, nil
	}
	end := offset + limit
	if end > total || limit == 0 {
		end = total
	}
	return all[offset:end], total, nil
}

func (s *inMemoryStore) Percentile(_ context.Context, mode, userID string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := s.sortedEntries(mode)
	var current storage.Entry
	found := false
	for _, entry := range all {
		if entry.UserID == userID {
			current = entry
			found = true
			break
		}
	}
	if !found {
		return 0, storage.ErrUserNotFound
	}
	totalEligible := len(all) - 1
	if totalEligible <= 0 {
		return 0, nil
	}
	worse := 0
	for _, entry := range all {
		if entry.UserID == userID {
			continue
		}
		if ranksBelow(entry, current) {
			worse++
		}
	}
	return float64(worse) / float64(totalEligible) * 100, nil
}

func (s *inMemoryStore) sortedEntries(mode string) []storage.Entry {
	modeEntries := s.entries[mode]
	all := make([]storage.Entry, 0, len(modeEntries))
	for _, entry := range modeEntries {
		all = append(all, entry)
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Score != all[j].Score {
			return all[i].Score > all[j].Score
		}
		if all[i].Accuracy != all[j].Accuracy {
			return all[i].Accuracy > all[j].Accuracy
		}
		return all[i].FinishedAt.Before(all[j].FinishedAt)
	})
	return all
}

func ranksBelow(candidate, reference storage.Entry) bool {
	if candidate.Score != reference.Score {
		return candidate.Score < reference.Score
	}
	if candidate.Accuracy != reference.Accuracy {
		return candidate.Accuracy < reference.Accuracy
	}
	return candidate.FinishedAt.After(reference.FinishedAt)
}

func seedStore() *inMemoryStore {
	store := newInMemoryStore()
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	_ = store.AddSession(context.Background(), "session-1", "user-a", "vanilla", base)
	_ = store.AddSession(context.Background(), "session-2", "user-b", "vanilla", base.Add(time.Hour))
	_ = store.UpsertEntry(context.Background(), "vanilla", storage.Entry{UserID: "user-a", Score: 100, TotalAttempts: 10, Accuracy: 0.8, FinishedAt: base.Add(2 * time.Hour)})
	_ = store.UpsertEntry(context.Background(), "vanilla", storage.Entry{UserID: "user-b", Score: 90, TotalAttempts: 12, Accuracy: 0.75, FinishedAt: base.Add(3 * time.Hour)})
	_ = store.UpsertEntry(context.Background(), "vanilla", storage.Entry{UserID: "user-c", Score: 100, TotalAttempts: 8, Accuracy: 0.9, FinishedAt: base.Add(time.Hour)})
	return store
}

func setupTestServer(store storage.LeaderboardStore) (leaderboardpb.LeaderboardServiceClient, func()) {
	srv := grpc.NewServer()
	svc := NewLeaderboardServiceServer(store)
	leaderboardpb.RegisterLeaderboardServiceServer(srv, svc)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	go srv.Serve(lis)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	client := leaderboardpb.NewLeaderboardServiceClient(conn)
	return client, func() {
		conn.Close()
		srv.Stop()
	}
}

func TestLeaderboardService_GetLeaderboard(t *testing.T) {
	client, cleanup := setupTestServer(seedStore())
	defer cleanup()

	resp, err := client.GetLeaderboard(context.Background(), &leaderboardpb.GetLeaderboardRequest{
		Mode: "vanilla", Limit: 2, Offset: 0,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(3), resp.Total)
	require.Len(t, resp.Entries, 2)
	assert.Equal(t, int32(1), resp.Entries[0].Rank)
	assert.Equal(t, "user-c", resp.Entries[0].UserId)
	assert.Equal(t, int32(2), resp.Entries[1].Rank)
	assert.Equal(t, "user-a", resp.Entries[1].UserId)
}

func TestLeaderboardService_GetPercentile(t *testing.T) {
	client, cleanup := setupTestServer(seedStore())
	defer cleanup()

	t.Run("returns correct percentile", func(t *testing.T) {
		resp, err := client.GetPercentile(context.Background(), &leaderboardpb.GetPercentileRequest{
			Mode: "vanilla", UserId: "user-a",
		})
		require.NoError(t, err)
		assert.InDelta(t, 50.0, resp.Percentile, 0.001)
	})

	t.Run("top rank gets 100 percentile", func(t *testing.T) {
		resp, err := client.GetPercentile(context.Background(), &leaderboardpb.GetPercentileRequest{
			Mode: "vanilla", UserId: "user-c",
		})
		require.NoError(t, err)
		assert.InDelta(t, 100.0, resp.Percentile, 0.001)
	})

	t.Run("bottom rank gets 0 percentile", func(t *testing.T) {
		resp, err := client.GetPercentile(context.Background(), &leaderboardpb.GetPercentileRequest{
			Mode: "vanilla", UserId: "user-b",
		})
		require.NoError(t, err)
		assert.InDelta(t, 0.0, resp.Percentile, 0.001)
	})
}

func TestLeaderboardService_GetPercentile_UserNotFound(t *testing.T) {
	client, cleanup := setupTestServer(seedStore())
	defer cleanup()

	_, err := client.GetPercentile(context.Background(), &leaderboardpb.GetPercentileRequest{
		Mode: "vanilla", UserId: "user-z",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLeaderboardService_AddSession(t *testing.T) {
	store := newInMemoryStore()
	client, cleanup := setupTestServer(store)
	defer cleanup()

	_, err := client.AddSession(context.Background(), &leaderboardpb.AddSessionRequest{
		SessionId: "session-3", UserId: "user-x", Mode: "blind",
	})
	require.NoError(t, err)

	sessions, err := store.ListActiveSessions(context.Background())
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Equal(t, "session-3", sessions[0].ID)
}

func TestLeaderboardService_RemoveSession(t *testing.T) {
	store := newInMemoryStore()
	_ = store.AddSession(context.Background(), "session-1", "user-a", "vanilla", time.Now())
	_ = store.AddSession(context.Background(), "session-2", "user-b", "vanilla", time.Now())

	client, cleanup := setupTestServer(store)
	defer cleanup()

	_, err := client.RemoveSession(context.Background(), &leaderboardpb.RemoveSessionRequest{SessionId: "session-1"})
	require.NoError(t, err)

	sessions, err := store.ListActiveSessions(context.Background())
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Equal(t, "session-2", sessions[0].ID)
}

func TestLeaderboardService_UpsertEntry(t *testing.T) {
	store := newInMemoryStore()
	client, cleanup := setupTestServer(store)
	defer cleanup()

	_, err := client.UpsertEntry(context.Background(), &leaderboardpb.UpsertEntryRequest{
		Mode: "vanilla", UserId: "user-a", Score: 100, TotalAttempts: 10, Accuracy: 0.8,
	})
	require.NoError(t, err)

	entries, total, err := store.ListEntries(context.Background(), "vanilla", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, entries, 1)
	assert.Equal(t, "user-a", entries[0].UserID)
	assert.Equal(t, 100, entries[0].Score)
}

func TestLeaderboardService_GetLeaderboard_MissingMode(t *testing.T) {
	client, cleanup := setupTestServer(newInMemoryStore())
	defer cleanup()

	_, err := client.GetLeaderboard(context.Background(), &leaderboardpb.GetLeaderboardRequest{})
	require.Error(t, err)
}

func TestLeaderboardService_GetPercentile_MissingParams(t *testing.T) {
	client, cleanup := setupTestServer(newInMemoryStore())
	defer cleanup()

	_, err := client.GetPercentile(context.Background(), &leaderboardpb.GetPercentileRequest{})
	require.Error(t, err)

	_, err = client.GetPercentile(context.Background(), &leaderboardpb.GetPercentileRequest{Mode: "vanilla"})
	require.Error(t, err)
}

func TestLeaderboardService_AddSession_MissingSessionID(t *testing.T) {
	client, cleanup := setupTestServer(newInMemoryStore())
	defer cleanup()

	_, err := client.AddSession(context.Background(), &leaderboardpb.AddSessionRequest{})
	require.Error(t, err)
}

func TestLeaderboardService_RemoveSession_MissingSessionID(t *testing.T) {
	client, cleanup := setupTestServer(newInMemoryStore())
	defer cleanup()

	_, err := client.RemoveSession(context.Background(), &leaderboardpb.RemoveSessionRequest{})
	require.Error(t, err)
}

func TestLeaderboardService_UpsertEntry_MissingFields(t *testing.T) {
	client, cleanup := setupTestServer(newInMemoryStore())
	defer cleanup()

	_, err := client.UpsertEntry(context.Background(), &leaderboardpb.UpsertEntryRequest{})
	require.Error(t, err)

	_, err = client.UpsertEntry(context.Background(), &leaderboardpb.UpsertEntryRequest{Mode: "vanilla"})
	require.Error(t, err)
}
