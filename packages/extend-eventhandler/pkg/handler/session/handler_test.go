package session

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	eventsv1 "extend-event-handler/pkg/pb/atoyr/events/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeLeaderboard struct {
	sessions []string
	entries  map[string]struct {
		userID        string
		score         int
		totalAttempts int
		accuracy      float64
		finishedAt    time.Time
	}
}

func newFakeLeaderboard() *fakeLeaderboard {
	return &fakeLeaderboard{
		entries: make(map[string]struct {
			userID        string
			score         int
			totalAttempts int
			accuracy      float64
			finishedAt    time.Time
		}),
	}
}

func (f *fakeLeaderboard) AddSession(_ context.Context, sessionID, _, _ string, _ time.Time) error {
	f.sessions = append(f.sessions, sessionID)
	return nil
}

func (f *fakeLeaderboard) RemoveSession(_ context.Context, sessionID string) error {
	filtered := f.sessions[:0]
	for _, s := range f.sessions {
		if s != sessionID {
			filtered = append(filtered, s)
		}
	}
	f.sessions = filtered
	return nil
}

func (f *fakeLeaderboard) UpsertEntry(_ context.Context, mode, userID string, score int, totalAttempts int, accuracy float64, finishedAt time.Time) error {
	f.entries[mode+"/"+userID] = struct {
		userID        string
		score         int
		totalAttempts int
		accuracy      float64
		finishedAt    time.Time
	}{
		userID:        userID,
		score:         score,
		totalAttempts: totalAttempts,
		accuracy:      accuracy,
		finishedAt:    finishedAt,
	}
	return nil
}

func marshalPayload(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func TestHandler_SessionStarted(t *testing.T) {
	lb := newFakeLeaderboard()
	h := NewSessionHandler(lb)

	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	envelope := &eventsv1.EventEnvelope{
		EventName: eventSessionStarted,
		Payload: marshalPayload(t, map[string]any{
			"session_id": "session-1",
			"user_id":    "user-a",
			"mode":       "vanilla",
			"started_at": now.Format(time.RFC3339Nano),
		}),
	}

	_, err := h.OnEvent(context.Background(), envelope)
	require.NoError(t, err)
	assert.Contains(t, lb.sessions, "session-1")
}

func TestHandler_SessionFinished(t *testing.T) {
	lb := newFakeLeaderboard()
	lb.sessions = append(lb.sessions, "session-1")
	h := NewSessionHandler(lb)

	finishedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	envelope := &eventsv1.EventEnvelope{
		EventName: eventSessionFinished,
		Payload: marshalPayload(t, map[string]any{
			"session_id":     "session-1",
			"user_id":        "user-a",
			"mode":           "vanilla",
			"score":          100,
			"total_attempts": 10,
			"accuracy":       0.85,
			"finished_at":    finishedAt.Format(time.RFC3339Nano),
		}),
	}

	_, err := h.OnEvent(context.Background(), envelope)
	require.NoError(t, err)

	assert.NotContains(t, lb.sessions, "session-1")
	entry, ok := lb.entries["vanilla/user-a"]
	require.True(t, ok)
	assert.Equal(t, 100, entry.score)
	assert.Equal(t, 10, entry.totalAttempts)
	assert.InDelta(t, 0.85, entry.accuracy, 0.001)
	assert.Equal(t, finishedAt, entry.finishedAt)
}

func TestHandler_RoundFinished(t *testing.T) {
	lb := newFakeLeaderboard()
	h := NewSessionHandler(lb)

	envelope := &eventsv1.EventEnvelope{
		EventName: eventRoundFinished,
		Payload: marshalPayload(t, map[string]any{
			"session_id":     "session-2",
			"user_id":        "user-b",
			"mode":           "blind",
			"score":          80,
			"total_attempts": 8,
			"accuracy":       0.7,
			"finished_at":    time.Now().Format(time.RFC3339Nano),
		}),
	}

	_, err := h.OnEvent(context.Background(), envelope)
	require.NoError(t, err)

	entry, ok := lb.entries["blind/user-b"]
	require.True(t, ok)
	assert.Equal(t, 80, entry.score)
}

func TestHandler_UnknownEvent(t *testing.T) {
	lb := newFakeLeaderboard()
	h := NewSessionHandler(lb)

	envelope := &eventsv1.EventEnvelope{
		EventName: "atoyr.session.unknown",
		Payload:   []byte("{}"),
	}

	_, err := h.OnEvent(context.Background(), envelope)
	require.Error(t, err)
}

func TestHandler_InvalidPayload(t *testing.T) {
	lb := newFakeLeaderboard()
	h := NewSessionHandler(lb)

	envelope := &eventsv1.EventEnvelope{
		EventName: eventSessionStarted,
		Payload:   []byte("{invalid"),
	}

	_, err := h.OnEvent(context.Background(), envelope)
	require.Error(t, err)
}
