package services

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"atoyr/server/internal/core"
	"atoyr/server/internal/middleware"
	"atoyr/server/internal/models"
	"atoyr/server/internal/testutils"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestStatsService_GetStats(t *testing.T) {
	db := testutils.SetupTestDB(t)
	metricsCollector := middleware.NewMetricsCollector()
	service := NewStatsService(db, metricsCollector)

	now := time.Now()
	today := now.Truncate(24 * time.Hour)

	sessions := []models.SessionEntity{
		{
			ID: "session-today-1", CreatedAt: today.Add(1 * time.Hour), UpdatedAt: today.Add(1 * time.Hour), EndsAt: today.Add(2 * time.Hour),
			Phase: core.SessionPhaseFinished, Score: 100, TotalAttempts: 10, DurationSeconds: 60,
			CorrectAttemptTimestamps: models.JSON{}, UsedWords: pq.StringArray{}, WordDefinitions: pq.StringArray{}, UsedItemIDs: pq.StringArray{},
		},
		{
			ID: "session-today-2", CreatedAt: today.Add(2 * time.Hour), UpdatedAt: today.Add(2 * time.Hour), EndsAt: today.Add(3 * time.Hour),
			Phase: core.SessionPhasePlaying, Score: 50, TotalAttempts: 5, DurationSeconds: 60,
			CorrectAttemptTimestamps: models.JSON{}, UsedWords: pq.StringArray{}, WordDefinitions: pq.StringArray{}, UsedItemIDs: pq.StringArray{},
		},
		{
			ID: "session-yesterday", CreatedAt: today.AddDate(0, 0, -1).Add(12 * time.Hour), UpdatedAt: today.AddDate(0, 0, -1).Add(12 * time.Hour), EndsAt: today.AddDate(0, 0, -1).Add(13 * time.Hour),
			Phase: core.SessionPhaseFinished, Score: 75, TotalAttempts: 8, DurationSeconds: 60,
			CorrectAttemptTimestamps: models.JSON{}, UsedWords: pq.StringArray{}, WordDefinitions: pq.StringArray{}, UsedItemIDs: pq.StringArray{},
		},
		{
			ID: "session-last-week", CreatedAt: today.AddDate(0, 0, -8), UpdatedAt: today.AddDate(0, 0, -8), EndsAt: today.AddDate(0, 0, -8).Add(1 * time.Hour),
			Phase: core.SessionPhaseFinished, Score: 25, TotalAttempts: 3, DurationSeconds: 60,
			CorrectAttemptTimestamps: models.JSON{}, UsedWords: pq.StringArray{}, WordDefinitions: pq.StringArray{}, UsedItemIDs: pq.StringArray{},
		},
	}

	for _, s := range sessions {
		err := db.Create(&s).Error
		assert.NoError(t, err)
	}

	metricsCollector.RequestCount.Store(42)
	metricsCollector.ErrorCount4xx.Store(3)
	metricsCollector.ErrorCount5xx.Store(1)
	metricsCollector.ResponseTimes = append(metricsCollector.ResponseTimes,
		100*time.Millisecond, 200*time.Millisecond, 300*time.Millisecond,
	)

	stats, err := service.GetStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	assert.Equal(t, int64(2), stats.SessionsToday)
	assert.Equal(t, int64(3), stats.SessionsThisWeek)
	assert.Equal(t, int64(3), stats.SessionsThisMonth)
	assert.Equal(t, int64(1), stats.ActiveGames)

	assert.Equal(t, int64(42), stats.TotalRequests)
	assert.Equal(t, float32(3), stats.ErrorRate4xx)
	assert.Equal(t, float32(1), stats.ErrorRate5xx)
	assert.Equal(t, int64(200), stats.ResponseTimeP50)
	assert.Equal(t, int64(300), stats.ResponseTimeP95)
	assert.Equal(t, int64(300), stats.ResponseTimeP99)
	assert.Greater(t, stats.MemoryUsage, uint64(0))
	assert.GreaterOrEqual(t, stats.Uptime, int64(0))
}

func TestStatsService_GetStats_Empty(t *testing.T) {
	db := testutils.SetupTestDB(t)
	metricsCollector := middleware.NewMetricsCollector()
	service := NewStatsService(db, metricsCollector)

	stats, err := service.GetStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	assert.Equal(t, int64(0), stats.SessionsToday)
	assert.Equal(t, int64(0), stats.SessionsThisWeek)
	assert.Equal(t, int64(0), stats.SessionsThisMonth)
	assert.Equal(t, int64(0), stats.ActiveGames)
	assert.Equal(t, int64(0), stats.TotalRequests)
	assert.Equal(t, float32(0), stats.ErrorRate4xx)
	assert.Equal(t, float32(0), stats.ErrorRate5xx)
}

func TestStatsService_GetTimeSeries(t *testing.T) {
	db := testutils.SetupTestDB(t)
	metricsCollector := middleware.NewMetricsCollector()
	service := NewStatsService(db, metricsCollector)

	now := time.Now()
	snapshots := []models.MetricSnapshot{
		{Timestamp: now.Add(-30 * time.Minute).UTC(), RequestCount: 100, ErrorCount4xx: 2, ErrorCount5xx: 1, ResponseTimeP50: 50, ResponseTimeP95: 100, ResponseTimeP99: 200, MemoryUsageMb: 3.1},
		{Timestamp: now.Add(-20 * time.Minute).UTC(), RequestCount: 200, ErrorCount4xx: 1, ErrorCount5xx: 0, ResponseTimeP50: 60, ResponseTimeP95: 110, ResponseTimeP99: 210, MemoryUsageMb: 3.2},
		{Timestamp: now.Add(-10 * time.Minute).UTC(), RequestCount: 150, ErrorCount4xx: 0, ErrorCount5xx: 1, ResponseTimeP50: 55, ResponseTimeP95: 105, ResponseTimeP99: 205, MemoryUsageMb: 3.3},
	}
	for _, s := range snapshots {
		err := db.Create(&s).Error
		assert.NoError(t, err)
	}

	result, err := service.GetTimeSeries("1h", "5m")
	assert.NoError(t, err)
	assert.Equal(t, "1h", result.Period)
	assert.Equal(t, "5m", result.Granularity)
	assert.Len(t, result.Data, 12)

	sort.Slice(result.Data, func(i, j int) bool {
		return result.Data[i].Timestamp.Before(result.Data[j].Timestamp)
	})

	findInSlice := func(fn func(s TimeSeriesPoint) bool) *TimeSeriesPoint {
		for _, s := range result.Data {
			if fn(s) {
				return &s
			}
		}

		return nil
	}

	tsWithRequestCount100 := findInSlice(func(s TimeSeriesPoint) bool {
		return s.RequestCount == 100
	})
	assert.InDelta(t, float32(3)/float32(100), tsWithRequestCount100.ErrorRate, 0.001)
	assert.Equal(t, float32(3.1), tsWithRequestCount100.MemoryUsage)

	tsWithRequestCount200 := findInSlice(func(s TimeSeriesPoint) bool {
		return s.RequestCount == 200
	})
	assert.InDelta(t, float32(1)/float32(200), tsWithRequestCount200.ErrorRate, 0.001)
	assert.Equal(t, float32(3.2), tsWithRequestCount200.MemoryUsage)

	tsWithRequestCount300 := findInSlice(func(s TimeSeriesPoint) bool {
		return s.RequestCount == 150
	})
	assert.InDelta(t, float32(1)/float32(150), tsWithRequestCount300.ErrorRate, 0.001)
	assert.Equal(t, float32(3.3), tsWithRequestCount300.MemoryUsage)
}

func TestStatsService_GetTimeSeries_Empty(t *testing.T) {
	db := testutils.SetupTestDB(t)
	metricsCollector := middleware.NewMetricsCollector()
	service := NewStatsService(db, metricsCollector)

	result, err := service.GetTimeSeries("1h", "5m")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Data, 12)

	for _, item := range result.Data {
		assert.Zero(t, item.ErrorRate)
		assert.Zero(t, item.MemoryUsage)
		assert.Zero(t, item.RequestCount)
		assert.Zero(t, item.ResponseTimeP50)
		assert.Zero(t, item.ResponseTimeP95)
		assert.Zero(t, item.ResponseTimeP99)
	}

	result, err = service.GetTimeSeries("24h", "1h")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Data, 24)

	for _, item := range result.Data {
		assert.Zero(t, item.ErrorRate)
		assert.Zero(t, item.MemoryUsage)
		assert.Zero(t, item.RequestCount)
		assert.Zero(t, item.ResponseTimeP50)
		assert.Zero(t, item.ResponseTimeP95)
		assert.Zero(t, item.ResponseTimeP99)
	}
}

func TestStatsService_GetTimeSeries_InvalidPeriod(t *testing.T) {
	db := testutils.SetupTestDB(t)
	metricsCollector := middleware.NewMetricsCollector()
	service := NewStatsService(db, metricsCollector)

	_, err := service.GetTimeSeries("invalid", "5m")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid period")
}

func TestAggregateSnapshots(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, &time.Location{})
	snapshots := []models.MetricSnapshot{
		{Timestamp: now.Add(-3 * time.Minute).UTC(), RequestCount: 100, ErrorCount4xx: 2, ErrorCount5xx: 1, ResponseTimeP50: 50, ResponseTimeP95: 100, ResponseTimeP99: 200, MemoryUsageMb: 3.1},
		{Timestamp: now.Add(-8 * time.Minute).UTC(), RequestCount: 100, ErrorCount4xx: 2, ErrorCount5xx: 1, ResponseTimeP50: 50, ResponseTimeP95: 100, ResponseTimeP99: 200, MemoryUsageMb: 3.1},
		{Timestamp: now.Add(-12 * time.Minute).UTC(), RequestCount: 100, ErrorCount4xx: 2, ErrorCount5xx: 1, ResponseTimeP50: 50, ResponseTimeP95: 100, ResponseTimeP99: 200, MemoryUsageMb: 3.1},
		{Timestamp: now.Add(-15 * time.Minute).UTC(), RequestCount: 100, ErrorCount4xx: 2, ErrorCount5xx: 1, ResponseTimeP50: 50, ResponseTimeP95: 100, ResponseTimeP99: 200, MemoryUsageMb: 3.1},
		{Timestamp: now.Add(-29 * time.Minute).UTC(), RequestCount: 100, ErrorCount4xx: 2, ErrorCount5xx: 1, ResponseTimeP50: 50, ResponseTimeP95: 100, ResponseTimeP99: 200, MemoryUsageMb: 3.1},
		{Timestamp: now.Add(-41 * time.Minute).UTC(), RequestCount: 100, ErrorCount4xx: 2, ErrorCount5xx: 1, ResponseTimeP50: 50, ResponseTimeP95: 100, ResponseTimeP99: 200, MemoryUsageMb: 3.1},
	}

	result, err := aggregateSnapshots(snapshots, now.Add(-1*time.Hour), now, time.Duration(5*time.Minute))
	assert.NoError(t, err)
	assert.Len(t, result, 12)

	// Every point should be before the next one.
	for i, p := range result {
		if i+1 == len(result) {
			break
		}

		assert.True(t, p.Timestamp.Before(result[i+1].Timestamp), fmt.Sprintf("%v %v\n", p.Timestamp, result[i+1].Timestamp))
	}

	result, err = aggregateSnapshots(snapshots, now.Add(-1*time.Hour), now, time.Minute)
	assert.NoError(t, err)
	assert.Len(t, result, 60)

	for i, p := range result {
		if i+1 == len(result) {
			break
		}

		assert.True(t, p.Timestamp.Before(result[i+1].Timestamp), fmt.Sprintf("%v %v\n", p.Timestamp, result[i+1].Timestamp))
	}
}
