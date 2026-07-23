package services

import (
	"fmt"

	"atoyr/server/internal/services/domainmodels"
)

type LeaderboardService struct {
	sessionService *SessionService
	leaderboard    LeaderboardReadClient
}

func NewLeaderboardService(sessionService *SessionService, leaderboardClient LeaderboardReadClient) *LeaderboardService {
	return &LeaderboardService{
		sessionService: sessionService,
		leaderboard:    leaderboardClient,
	}
}

type LeaderboardEntry struct {
	ID            string  `json:"id"`
	Rank          int32   `json:"rank"`
	Score         int32   `json:"score"`
	TotalAttempts int32   `json:"totalAttempts"`
	Accuracy      float32 `json:"accuracy"`
	Timestamp     int64   `json:"timestamp"`
}

func (l *LeaderboardService) GetLeaderboard(mode string, limit, offset int) ([]LeaderboardEntry, error) {
	if l.leaderboard == nil {
		return nil, fmt.Errorf("leaderboard backend not configured")
	}
	entries, _, err := l.leaderboard.GetLeaderboard(mode, limit, offset)
	return entries, err
}

func (l *LeaderboardService) GetTotalEntries(mode string) (int64, error) {
	if l.leaderboard == nil {
		return 0, fmt.Errorf("leaderboard backend not configured")
	}
	return l.leaderboard.GetTotalEntries(mode)
}

func (l *LeaderboardService) GetPercentile(sessionId, mode string) (float32, error) {
	if l.leaderboard == nil {
		return 0, fmt.Errorf("leaderboard backend not configured")
	}
	session, err := l.sessionService.FindByID(sessionId)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch session: %w", err)
	}
	if session.UserID == "" {
		return 0, fmt.Errorf("session has no user id")
	}

	percentile, err := l.leaderboard.GetPercentile(mode, session.UserID)
	if err != nil {
		return 0, err
	}
	return float32(percentile), nil
}

func (l *LeaderboardService) UpsertEntry(session *domainmodels.SessionDomain) error {
	if l.leaderboard == nil {
		return fmt.Errorf("leaderboard backend not configured")
	}
	if session.UserID == "" {
		return nil
	}
	return l.leaderboard.UpsertEntry(session.Mode, session.UserID, session.Score, session.TotalAttempts, session.Accuracy, session.EndsAt.UnixMilli())
}
