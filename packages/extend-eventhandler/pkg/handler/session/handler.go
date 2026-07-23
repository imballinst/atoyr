package session

import (
	"context"
	"encoding/json"
	"time"

	eventsv1 "extend-event-handler/pkg/pb/atoyr/events/v1"
	"extend-event-handler/pkg/leaderboard"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	eventSessionStarted = "atoyr.session.started"
	eventSessionFinished = "atoyr.session.finished"
	eventRoundFinished   = "atoyr.round.finished"
)

type SessionHandler struct {
	eventsv1.UnimplementedAtoyrEventServiceServer
	leaderboard leaderboard.Client
}

func NewSessionHandler(lb leaderboard.Client) *SessionHandler {
	return &SessionHandler{leaderboard: lb}
}

func (h *SessionHandler) OnEvent(ctx context.Context, envelope *eventsv1.EventEnvelope) (*emptypb.Empty, error) {
	switch envelope.EventName {
	case eventSessionStarted:
		return h.handleSessionStarted(ctx, envelope.Payload)
	case eventSessionFinished, eventRoundFinished:
		return h.handleSessionFinished(ctx, envelope.Payload)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown event: %s", envelope.EventName)
	}
}

func (h *SessionHandler) handleSessionStarted(ctx context.Context, payload []byte) (*emptypb.Empty, error) {
	var evt struct {
		SessionID string    `json:"session_id"`
		UserID    string    `json:"user_id"`
		Mode      string    `json:"mode"`
		StartedAt time.Time `json:"started_at"`
	}
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid session.started payload: %v", err)
	}

	return &emptypb.Empty{}, h.leaderboard.AddSession(ctx, evt.SessionID, evt.UserID, evt.Mode, evt.StartedAt)
}

func (h *SessionHandler) handleSessionFinished(ctx context.Context, payload []byte) (*emptypb.Empty, error) {
	var evt struct {
		SessionID     string    `json:"session_id"`
		UserID        string    `json:"user_id"`
		Mode          string    `json:"mode"`
		Score         int       `json:"score"`
		TotalAttempts int       `json:"total_attempts"`
		Accuracy      float64   `json:"accuracy"`
		FinishedAt    time.Time `json:"finished_at"`
	}
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid session.finished payload: %v", err)
	}

	if err := h.leaderboard.RemoveSession(ctx, evt.SessionID); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, h.leaderboard.UpsertEntry(ctx, evt.Mode, evt.UserID, evt.Score, evt.TotalAttempts, evt.Accuracy, evt.FinishedAt)
}
