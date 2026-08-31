package services

import (
	"fmt"
	"time"

	"atoyr/server/internal/core"
	"atoyr/server/internal/models"
	"atoyr/server/internal/utils"

	"gorm.io/gorm"
)

const (
	LeaderboardPeriodAlltime = "alltime"
	LeaderboardPeriodMonthly = "monthly"
)

type LeaderboardService struct {
	db  *gorm.DB
	Now func() time.Time
}

func NewLeaderboardService(db *gorm.DB) *LeaderboardService {
	return &LeaderboardService{db: db, Now: time.Now}
}

type LeaderboardEntry struct {
	ID            string  `json:"id"`
	Rank          int32   `json:"rank"`
	Score         int32   `json:"score"`
	TotalAttempts int32   `json:"totalAttempts"`
	Accuracy      float32 `json:"accuracy"`
	Timestamp     int64   `json:"timestamp"`
	Topic         string  `json:"topic"`
}

// periodFilter returns a SQL condition fragment and its args, restricting the
// eligible rows to the given period. For "alltime" it returns an empty
// fragment. For "monthly" it filters to the start of the current month using
// the service's clock so the test suite can mock time.
func (l *LeaderboardService) periodFilter(period string) (string, []any) {
	if period == LeaderboardPeriodMonthly {
		now := l.Now().UTC()
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return "ends_at >= ?", []any{startOfMonth}
	}
	return "", nil
}

func (l *LeaderboardService) GetLeaderboard(mode, topic, period string, limit, offset int) ([]LeaderboardEntry, error) {
	var results []models.SessionEntity
	query := l.db.Where("mode = ? AND topic = ? AND phase = ? AND score > 0", mode, topic, core.SessionPhaseFinished)
	if fragment, args := l.periodFilter(period); fragment != "" {
		query = query.Where(fragment, args...)
	}

	if err := query.
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
			Topic:         result.Topic,
		}
	}

	return entries, nil
}

func (l *LeaderboardService) GetTotalEntries(mode, topic, period string) (int64, error) {
	var totalEntries int64
	query := l.db.Model(&models.SessionEntity{}).
		Where("mode = ? AND topic = ? AND phase = ? AND score > 0", mode, topic, core.SessionPhaseFinished)
	if fragment, args := l.periodFilter(period); fragment != "" {
		query = query.Where(fragment, args...)
	}

	if err := query.Count(&totalEntries).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch leaderboard: %w", err)
	}

	return totalEntries, nil
}

func (l *LeaderboardService) GetPercentile(sessionId, mode, topic, period string, score int32) (float32, int32, error) {
	// 1. Fetch the current session's tiebreaker fields
	type sessionMeta struct {
		Accuracy float32
		EndsAt   time.Time
	}
	var meta sessionMeta
	if err := l.db.
		Raw("SELECT accuracy, ends_at FROM session_entities WHERE id = ?", sessionId).
		Scan(&meta).Error; err != nil {
		return 0, 0, fmt.Errorf("failed to fetch session meta: %w", err)
	}

	// 2. Total eligible
	var totalEligible int64
	eligibleQuery := l.db.Model(&models.SessionEntity{}).
		Where("phase = ? AND mode = ? AND topic = ? AND score > 0 AND id != ?",
			core.SessionPhaseFinished, mode, topic, sessionId)
	if fragment, args := l.periodFilter(period); fragment != "" {
		eligibleQuery = eligibleQuery.Where(fragment, args...)
	}
	if err := eligibleQuery.Count(&totalEligible).Error; err != nil {
		return 0, 0, fmt.Errorf("failed to fetch leaderboard: %w", err)
	}

	// 3. Count sessions that rank WORSE than the current one
	var totalBelowCurrentScore int64
	worseQuery := l.db.Model(&models.SessionEntity{}).
		Where(`phase = ? AND mode = ? AND topic = ? AND score > 0 AND id != ?
               AND (
                    score < ?
                    OR (score = ? AND accuracy < ?)
                    OR (score = ? AND accuracy = ? AND ends_at > ?)
               )`,
			core.SessionPhaseFinished, mode, topic, sessionId,
			score,
			score, meta.Accuracy,
			score, meta.Accuracy, meta.EndsAt)
	if fragment, args := l.periodFilter(period); fragment != "" {
		worseQuery = worseQuery.Where(fragment, args...)
	}
	if err := worseQuery.Count(&totalBelowCurrentScore).Error; err != nil {
		return 0, 0, fmt.Errorf("failed to fetch leaderboard: %w", err)
	}

	if totalEligible == 0 {
		return 0, 1, nil
	}

	percentile := (float32(totalBelowCurrentScore) / float32(totalEligible)) * 100
	rank := int32(totalEligible - totalBelowCurrentScore + 1)
	return percentile, rank, nil
}
