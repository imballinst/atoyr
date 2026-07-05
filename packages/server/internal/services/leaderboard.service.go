package services

import (
	"fmt"

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

func (l *LeaderboardService) GetLeaderboard(limit, offset int) ([]LeaderboardEntry, error) {
	var results []models.SessionEntity

	if err := l.db.
		Where("phase = ? AND score > 0", core.SessionPhaseFinished).
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

func (l *LeaderboardService) GetTotalEntries() (int64, error) {
	var totalEntries int64

	if err := l.db.
		Raw("SELECT COUNT(*) FROM session_entities WHERE phase = ?;", core.SessionPhaseFinished).Find(&totalEntries).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch leaderboard: %w", err)
	}

	return totalEntries, nil
}

func (l *LeaderboardService) GetPercentile(score int32) (float32, error) {
	var totalEligible int32
	var totalBelowCurrentScore int32

	if err := l.db.
		Raw("SELECT COUNT(*) FROM session_entities WHERE phase = ? AND score > 0 ORDER BY score DESC, accuracy DESC, ends_at ASC;", core.SessionPhaseFinished).Find(&totalEligible).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch leaderboard: %w", err)
	}
	if err := l.db.
		Raw("SELECT COUNT(*) FROM session_entities WHERE phase = ? AND score > 0 AND score < ? ORDER BY score DESC, accuracy DESC, ends_at ASC;", core.SessionPhaseFinished, score).Find(&totalBelowCurrentScore).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch leaderboard: %w", err)
	}

	percentile := (float32(totalBelowCurrentScore) / float32(totalEligible)) * 100
	return percentile, nil
}
