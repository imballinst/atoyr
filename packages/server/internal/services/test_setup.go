package services

import (
	"atoyr/server/internal/core"
	"atoyr/server/internal/testutils"
	"testing"

	"gorm.io/gorm"
)

type FakeLeaderboardClient struct {
	Entries       []LeaderboardEntry
	Total         int64
	Percentile    float64
	PercentileErr error
}

func NewFakeLeaderboardClient() *FakeLeaderboardClient {
	return &FakeLeaderboardClient{}
}

func (f *FakeLeaderboardClient) GetLeaderboard(_ string, limit, offset int) ([]LeaderboardEntry, int64, error) {
	if offset >= len(f.Entries) {
		return nil, f.Total, nil
	}
	end := offset + limit
	if end > len(f.Entries) {
		end = len(f.Entries)
	}
	return f.Entries[offset:end], f.Total, nil
}

func (f *FakeLeaderboardClient) GetPercentile(_, _ string) (float64, error) {
	return f.Percentile, f.PercentileErr
}

func (f *FakeLeaderboardClient) GetTotalEntries(_ string) (int64, error) {
	return f.Total, nil
}

func (f *FakeLeaderboardClient) UpsertEntry(_, _ string, score, attempts int32, accuracy float32, finishedAt int64) error {
	return nil
}

func initTestWordService() *WordService {
	return &WordService{
		Words: []WordDefinition{
			{Word: "hello", Definition: "a greeting"},
			{Word: "world", Definition: "the earth"},
			{Word: "apple", Definition: "a fruit"},
			{Word: "banana", Definition: "a yellow fruit"},
			{Word: "cherry", Definition: "a small red fruit"},
			{Word: "dragon", Definition: "a mythical creature"},
			{Word: "elephant", Definition: "a large mammal"},
			{Word: "forest", Definition: "a wooded area"},
			{Word: "guitar", Definition: "a stringed instrument"},
			{Word: "horizon", Definition: "where earth meets sky"},
		},
	}
}

type TestValues struct {
	DB          *gorm.DB
	Word        *WordService
	Session     *SessionService
	Leaderboard *LeaderboardService
	Game        *GameService
	Store       *core.SessionStore
}

func initTestServices(t *testing.T) TestValues {
	db := testutils.SetupTestDB(t)

	ws := initTestWordService()
	ss := NewSessionService(db)
	ls := NewLeaderboardService(ss, NewFakeLeaderboardClient())
	store := core.NewSessionStore()
	gs := NewGameService(ss, store, ws, ls, &AGSSyncService{}, testutils.TestSessionOptions)

	return TestValues{db, ws, ss, ls, gs, store}
}
