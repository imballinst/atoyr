package services

import (
	"fmt"
	"time"

	"atoyr/server/internal/core"
	"atoyr/server/internal/models"
	"atoyr/server/internal/utils"

	"gorm.io/gorm"
)

type LeaderboardService struct {
	db *gorm.DB
}

func NewLeaderboardService(db *gorm.DB) *LeaderboardService {
	return &LeaderboardService{db: db}
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
	var results []models.SessionEntity

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
	var totalEntries int64

	if err := l.db.
		Raw("SELECT COUNT(*) FROM session_entities WHERE mode = ? AND phase = ? AND score > 0;", mode, core.SessionPhaseFinished).Find(&totalEntries).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch leaderboard: %w", err)
	}

	return totalEntries, nil
}

func (l *LeaderboardService) GetPercentile(sessionId, mode string, score int32) (float32, error) {
	// 1. Fetch the current session's tiebreaker fields
	type sessionMeta struct {
		Accuracy float32
		EndsAt   time.Time
	}
	var meta sessionMeta
	if err := l.db.
		Raw("SELECT accuracy, ends_at FROM session_entities WHERE id = ?", sessionId).
		Scan(&meta).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch session meta: %w", err)
	}

	// 2. Total eligible
	var totalEligible int32
	if err := l.db.
		Raw(`SELECT COUNT(*) FROM session_entities
             WHERE phase = ? AND mode = ? AND score > 0 AND id != ?`,
			core.SessionPhaseFinished, mode, sessionId).
		Scan(&totalEligible).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch leaderboard: %w", err)
	}

	// 3. Count sessions that rank WORSE than the current one
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
			score,
			score, meta.Accuracy,
			score, meta.Accuracy, meta.EndsAt).
		Scan(&totalBelowCurrentScore).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch leaderboard: %w", err)
	}

	if totalEligible == 0 {
		return 0, nil
	}

	percentile := (float32(totalBelowCurrentScore) / float32(totalEligible)) * 100
	return percentile, nil
}


