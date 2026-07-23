package eventhandler

import (
	"atoyr/extend/leaderboard"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStore struct {
	sessions []string
	entries  map[string]leaderboard.Entry
}

func newFakeStore() *fakeStore {
	return &fakeStore{entries: make(map[string]leaderboard.Entry)}
}

func (f *fakeStore) AddSession(id, userID, mode string, startedAt time.Time) error {
	f.sessions = append(f.sessions, id)
	return nil
}

func (f *fakeStore) RemoveSession(id string) error {
	filtered := f.sessions[:0]
	for _, session := range f.sessions {
		if session != id {
			filtered = append(filtered, session)
		}
	}
	f.sessions = filtered
	return nil
}

func (f *fakeStore) UpsertEntry(mode string, entry leaderboard.Entry) error {
	f.entries[mode+"/"+entry.UserID] = entry
	return nil
}

func TestHandler_SessionStarted(t *testing.T) {
	store := newFakeStore()
	handler := NewHandler(store)

	err := handler.HandleSessionStarted(SessionStarted{
		EventName: "atoyr.session.started",
		SessionID: "session-1",
		UserID:    "user-a",
		Mode:      "vanilla",
		StartedAt: time.Now(),
	})
	require.NoError(t, err)
	assert.Contains(t, store.sessions, "session-1")
}

func TestHandler_SessionFinished(t *testing.T) {
	store := newFakeStore()
	_ = store.AddSession("session-1", "user-a", "vanilla", time.Now())
	handler := NewHandler(store)

	finishedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	err := handler.HandleSessionFinished(SessionFinished{
		EventName:     "atoyr.session.finished",
		SessionID:     "session-1",
		UserID:        "user-a",
		Mode:          "vanilla",
		Score:         100,
		TotalAttempts: 10,
		Accuracy:      0.85,
		FinishedAt:    finishedAt,
	})
	require.NoError(t, err)

	assert.NotContains(t, store.sessions, "session-1")
	entry, ok := store.entries["vanilla/user-a"]
	require.True(t, ok)
	assert.Equal(t, 100, entry.Score)
	assert.Equal(t, 10, entry.TotalAttempts)
	assert.InDelta(t, 0.85, entry.Accuracy, 0.001)
	assert.Equal(t, finishedAt, entry.FinishedAt)
}

func TestHandler_RoundFinished(t *testing.T) {
	store := newFakeStore()
	handler := NewHandler(store)

	err := handler.HandleRoundFinished(RoundFinished{
		EventName:     "atoyr.round.finished",
		SessionID:     "session-2",
		UserID:        "user-b",
		Mode:          "blind",
		Score:         80,
		TotalAttempts: 8,
		Accuracy:      0.7,
		FinishedAt:    time.Now(),
	})
	require.NoError(t, err)

	entry, ok := store.entries["blind/user-b"]
	require.True(t, ok)
	assert.Equal(t, 80, entry.Score)
}
