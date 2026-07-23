package eventhandler

import (
	"atoyr/extend/leaderboard"
	"time"
)

type Store interface {
	AddSession(id, userID, mode string, startedAt time.Time) error
	RemoveSession(id string) error
	UpsertEntry(mode string, entry leaderboard.Entry) error
}

type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) HandleSessionStarted(evt SessionStarted) error {
	return h.store.AddSession(evt.SessionID, evt.UserID, evt.Mode, evt.StartedAt)
}

func (h *Handler) HandleSessionFinished(evt SessionFinished) error {
	if err := h.store.RemoveSession(evt.SessionID); err != nil {
		return err
	}

	return h.store.UpsertEntry(evt.Mode, leaderboard.Entry{
		UserID:        evt.UserID,
		Score:         evt.Score,
		TotalAttempts: evt.TotalAttempts,
		Accuracy:      evt.Accuracy,
		FinishedAt:    evt.FinishedAt,
	})
}

func (h *Handler) HandleRoundFinished(evt RoundFinished) error {
	return h.HandleSessionFinished(SessionFinished(evt))
}
