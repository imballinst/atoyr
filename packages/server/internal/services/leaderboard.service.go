package services

import (
	"fmt"

	"atoyr/server/internal/core"
	"atoyr/server/internal/models"

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
		Order("score DESC, accuracy DESC").
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
			Accuracy:      result.Accuracy,
			Timestamp:     result.EndsAt.UnixMilli(),
		}
	}

	return entries, nil
}

func (l *LeaderboardService) GetTopScores(limit int) ([]LeaderboardEntry, error) {
	return l.GetLeaderboard(limit, 0)
}

func (l *LeaderboardService) GetTotalEntries() (int64, error) {
	var count int64
	if err := l.db.Model(&models.SessionEntity{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count results: %w", err)
	}
	return count, nil
}
