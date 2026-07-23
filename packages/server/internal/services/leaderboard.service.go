package services

import (
	"fmt"
	"time"

	"atoyr/server/internal/core"
	"atoyr/server/internal/models"
	"atoyr/server/internal/services/domainmodels"
	"atoyr/server/internal/utils"

	"gorm.io/gorm"
)

type LeaderboardService struct {
	db                      *gorm.DB
	sessionService          *SessionService
	agsSyncService          *AGSSyncService
	extendLeaderboardClient *ExtendLeaderboardClient
}

func NewLeaderboardService(db *gorm.DB, sessionService *SessionService, agsSyncService *AGSSyncService, extendClient *ExtendLeaderboardClient) *LeaderboardService {
	return &LeaderboardService{
		db:                      db,
		sessionService:          sessionService,
		agsSyncService:          agsSyncService,
		extendLeaderboardClient: extendClient,
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
	if l.extendLeaderboardClient != nil {
		entries, _, err := l.extendLeaderboardClient.GetLeaderboard(mode, limit, offset)
		return entries, err
	}

	if l.agsSyncService != nil && l.agsSyncService.Enabled() {
		entries, _, err := l.agsSyncService.GetLeaderboard(mode, limit, offset)
		return entries, err
	}

	var results []models.LeaderboardSessionEntity

	if err := l.db.
		Where("mode = ? AND phase = ? AND score > 0", mode, core.SessionPhaseFinished).
		Order("score DESC, accuracy DESC, ends_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch leaderboard: %w", err)
	}

	entries := make([]LeaderboardEntry, len(results))
	for i, result := range results {
		entries[i] = LeaderboardEntry{
			ID:            result.ID,
			Rank:          int32(i + 1 + offset),
			Score:         result.Score,
			TotalAttempts: result.TotalAttempts,
			Accuracy:      utils.ToPercentage(result.Accuracy),
			Timestamp:     result.EndsAt.UnixMilli(),
		}
	}

	return entries, nil
}

func (l *LeaderboardService) GetTotalEntries(mode string) (int64, error) {
	if l.extendLeaderboardClient != nil {
		return l.extendLeaderboardClient.GetTotalEntries(mode)
	}

	if l.agsSyncService != nil && l.agsSyncService.Enabled() {
		return l.agsSyncService.CountLeaderboardEntries(mode)
	}

	var totalEntries int64

	if err := l.db.
		Raw("SELECT COUNT(*) FROM session_entities WHERE mode = ? AND phase = ? AND score > 0;", mode, core.SessionPhaseFinished).Find(&totalEntries).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch leaderboard: %w", err)
	}

	return totalEntries, nil
}

func (l *LeaderboardService) GetPercentile(sessionId, mode string) (float32, error) {
	if l.extendLeaderboardClient != nil {
		session, err := l.sessionService.FindByID(sessionId)
		if err != nil {
			return 0, fmt.Errorf("failed to fetch session: %w", err)
		}
		if session.UserID == "" {
			return 0, fmt.Errorf("session has no user id")
		}
		percentile, err := l.extendLeaderboardClient.GetPercentile(mode, session.UserID)
		if err != nil {
			return 0, err
		}
		return float32(percentile), nil
	}

	if l.agsSyncService != nil && l.agsSyncService.Enabled() {
		session, err := l.sessionService.FindByID(sessionId)
		if err != nil {
			return 0, fmt.Errorf("failed to fetch session: %w", err)
		}
		if session.UserID == "" {
			return 0, fmt.Errorf("session has no user id")
		}
		rank, total, err := l.agsSyncService.GetUserRank(session.UserID, mode)
		if err != nil {
			return 0, err
		}
		if total <= 1 {
			return 0, nil
		}
		return (float32(total-rank) / float32(total-1)) * 100, nil
	}

	// SQLite fallback: read the finished session summary from the leaderboard
	// entity, which shares the session_entities table with the minimal registry.
	type sessionMeta struct {
		Score    int32
		Accuracy float32
		EndsAt   time.Time
	}
	var meta sessionMeta
	if err := l.db.
		Raw("SELECT score, accuracy, ends_at FROM session_entities WHERE id = ?", sessionId).
		Scan(&meta).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch session meta: %w", err)
	}

	var totalEligible int32
	if err := l.db.
		Raw(`SELECT COUNT(*) FROM session_entities
             WHERE phase = ? AND mode = ? AND score > 0 AND id != ?`,
			core.SessionPhaseFinished, mode, sessionId).
		Scan(&totalEligible).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch leaderboard: %w", err)
	}

	var totalBelowCurrentScore int32
	if err := l.db.
		Raw(`SELECT COUNT(*) FROM session_entities
             WHERE phase = ? AND mode = ? AND score > 0 AND id != ?
               AND (
                     score < ?
                     OR (score = ? AND accuracy < ?)
                     OR (score = ? AND accuracy = ? AND ends_at > ?)
               )`,
			core.SessionPhaseFinished, mode, sessionId,
			meta.Score,
			meta.Score, meta.Accuracy,
			meta.Score, meta.Accuracy, meta.EndsAt).
		Scan(&totalBelowCurrentScore).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch leaderboard: %w", err)
	}

	if totalEligible == 0 {
		return 0, nil
	}

	percentile := (float32(totalBelowCurrentScore) / float32(totalEligible)) * 100
	return percentile, nil
}

// SaveFallback persists a finished session's gameplay summary to SQLite for the
// leaderboard fallback. It is a no-op when AGS is enabled because AGS is the
// source of truth for leaderboard data in that configuration.
func (l *LeaderboardService) SaveFallback(session *domainmodels.SessionDomain) error {
	if l.extendLeaderboardClient != nil {
		return nil
	}

	if l.agsSyncService != nil && l.agsSyncService.Enabled() {
		return nil
	}

	entity, err := domainmodels.ConvertSessionDomainToLeaderboardSession(session)
	if err != nil {
		return fmt.Errorf("failed to convert session to leaderboard entity: %w", err)
	}
	entity.UpdatedAt = time.Now()

	if err := l.db.Save(entity).Error; err != nil {
		return fmt.Errorf("failed to save leaderboard fallback: %w", err)
	}
	return nil
}
