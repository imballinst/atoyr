package services

import (
	"atoyr/server/internal/middleware"
	"atoyr/server/internal/models"
	"atoyr/server/internal/utils"
	"fmt"
	"maps"
	"slices"
	"time"

	"gorm.io/gorm"
)

const (
	ErrorInvalidPeriod = "invalid period"
)

type StatsService struct {
	db        *gorm.DB
	metrics   *middleware.MetricsCollector
	startTime time.Time
	Now       func() time.Time
}

type StatsResponse struct {
	// From SQLite
	TotalSessions     int64 `json:"totalSessions"`
	SessionsToday     int64 `json:"sessionsToday"`
	SessionsThisWeek  int64 `json:"sessionsThisWeek"`
	SessionsThisMonth int64 `json:"sessionsThisMonth"`
	ActiveGames       int64 `json:"activeGames"`

	// From in-memory metrics
	TotalRequests   int64   `json:"totalRequests"`
	ErrorRate4xx    float32 `json:"errorRate4xx"`
	ErrorRate5xx    float32 `json:"errorRate5xx"`
	ResponseTimeP50 int64   `json:"responseTimeP50"` // milliseconds
	ResponseTimeP95 int64   `json:"responseTimeP95"` // milliseconds
	ResponseTimeP99 int64   `json:"responseTimeP99"` // milliseconds

	// Server resources
	MemoryUsage uint64 `json:"memoryUsageMB"`
	Uptime      int64  `json:"uptime"` // seconds
}

type TimeSeriesResponse struct {
	Period      string            `json:"period"`
	Granularity string            `json:"granularity"`
	Data        []TimeSeriesPoint `json:"data"`
}

type TimeSeriesPoint struct {
	Timestamp       time.Time `json:"timestamp"`
	RequestCount    int64     `json:"requestCount"`
	ErrorRate       float32   `json:"errorRate"`
	ResponseTimeP50 int64     `json:"responseTimeP50"`
	ResponseTimeP95 int64     `json:"responseTimeP95"`
	ResponseTimeP99 int64     `json:"responseTimeP99"`
	MemoryUsage     float32   `json:"memoryUsage"`
}

func NewStatsService(db *gorm.DB, metricsCollector *middleware.MetricsCollector) *StatsService {
	return &StatsService{
		db:        db,
		metrics:   metricsCollector,
		startTime: time.Now(),
		Now:       time.Now,
	}
}

func (s *StatsService) GetStats() (*StatsResponse, error) {
	now := s.Now().UTC()
	today := now.Truncate(24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour).Truncate(24 * time.Hour)
	weekStart, weekEnd := getWeekStartAndEnd(now)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonthStart := monthStart.AddDate(0, 1, 0)

	var stats StatsResponse

	// Sessions today
	result := s.db.Model(&models.SessionEntity{}).
		Where("created_at >= ? AND created_at < ?", today, tomorrow).
		Count(&stats.SessionsToday)
	if result.Error != nil {
		return nil, result.Error
	}

	// Sessions this week
	result = s.db.Model(&models.SessionEntity{}).
		Where("created_at >= ? AND created_at < ?", weekStart, weekEnd).
		Count(&stats.SessionsThisWeek)
	if result.Error != nil {
		return nil, result.Error
	}

	// Sessions this month
	result = s.db.Model(&models.SessionEntity{}).
		Where("created_at >= ? AND created_at < ?", monthStart, nextMonthStart).
		Count(&stats.SessionsThisMonth)
	if result.Error != nil {
		return nil, result.Error
	}

	// Active games (not finished)
	result = s.db.Model(&models.SessionEntity{}).
		Where("phase != ?", "finished").
		Count(&stats.ActiveGames)
	if result.Error != nil {
		return nil, result.Error
	}

	stats.TotalRequests = s.metrics.RequestCount.Load()
	stats.ErrorRate4xx = float32(s.metrics.ErrorCount4xx.Load())
	stats.ErrorRate5xx = float32(s.metrics.ErrorCount5xx.Load())
	stats.ResponseTimeP50 = utils.GetTimePercentile(s.metrics.ResponseTimes, 50)
	stats.ResponseTimeP95 = utils.GetTimePercentile(s.metrics.ResponseTimes, 95)
	stats.ResponseTimeP99 = utils.GetTimePercentile(s.metrics.ResponseTimes, 99)
	stats.MemoryUsage = uint64(utils.GetMemoryUsage())
	stats.Uptime = int64(time.Since(s.startTime).Seconds())

	return &stats, nil
}

func (s *StatsService) GetTimeSeries(period, granularity string) (*TimeSeriesResponse, error) {
	// Parse period and granularity
	startTime, endTime, interval, err := s.parseTimeRange(period, granularity)
	if err != nil {
		return nil, err
	}

	var snapshots []models.MetricSnapshot
	err = s.db.Where("timestamp >= ?", startTime).
		Order("timestamp ASC").
		Find(&snapshots).Error
	if err != nil {
		return nil, err
	}

	// Aggregate snapshots into time series points
	data, err := aggregateSnapshots(snapshots, startTime, endTime, interval)
	if err != nil {
		return nil, err
	}

	return &TimeSeriesResponse{
		Period:      period,
		Granularity: granularity,
		Data:        data,
	}, nil
}

func (s *StatsService) parseTimeRange(period, granularity string) (time.Time, time.Time, time.Duration, error) {
	now := s.Now().UTC()

	var startTime time.Time

	switch period {
	case "1h":
		startTime = now.Add(-1 * time.Hour)
	case "24h":
		startTime = now.Add(-24 * time.Hour)
	case "7d":
		startTime = now.AddDate(0, 0, -7)
	case "1M":
		startTime = now.AddDate(0, -1, 0)
	default:
		return time.Time{}, time.Time{}, 0, fmt.Errorf("%s: %s", ErrorInvalidPeriod, period)
	}

	var interval time.Duration
	switch granularity {
	case "1m":
		interval = time.Minute
	case "5m":
		interval = 5 * time.Minute
	case "1h":
		interval = time.Hour
	default:
		// Auto-select based on period
		switch period {
		case "1h":
			interval = time.Minute
		case "24h":
			interval = time.Hour
		case "7d", "1M":
			interval = 24 * time.Hour
		}
	}

	return startTime, now, interval, nil
}

func aggregateSnapshots(snapshots []models.MetricSnapshot, startTime, endTime time.Time, interval time.Duration) ([]TimeSeriesPoint, error) {
	grouped := map[string][]models.MetricSnapshot{}
	t := startTime

	for t.Before(endTime) && !t.Equal(endTime) {
		truncated := t.Truncate(interval)
		key := truncated.Format(time.RFC3339)

		grouped[key] = []models.MetricSnapshot{}
		t = t.Add(interval)
	}

	for _, snapshot := range snapshots {
		truncated := snapshot.Timestamp.Truncate(interval)
		key := truncated.Format(time.RFC3339)

		// Since every interval is already "populated" above, this should be guaranteed to exist.
		grouped[key] = append(grouped[key], snapshot)
	}

	keys := slices.Sorted((maps.Keys(grouped)))

	result := []TimeSeriesPoint{}

	for _, key := range keys {
		snapshotGroup := grouped[key]

		t, err := time.Parse(time.RFC3339, key)
		if err != nil {
			return result, err
		}

		requestCount := int64(0)
		errorCount := int64(0)
		p50 := []time.Duration{}
		p95 := []time.Duration{}
		p99 := []time.Duration{}
		totalMemoryUsage := float64(0)

		for _, snapshotItem := range snapshotGroup {
			requestCount += snapshotItem.RequestCount
			errorCount += snapshotItem.ErrorCount4xx + snapshotItem.ErrorCount5xx
			p50 = append(p50, time.Duration(snapshotItem.ResponseTimeP50)*time.Millisecond)
			p95 = append(p95, time.Duration(snapshotItem.ResponseTimeP95)*time.Millisecond)
			p99 = append(p99, time.Duration(snapshotItem.ResponseTimeP99)*time.Millisecond)
			totalMemoryUsage += snapshotItem.MemoryUsageMb
		}

		errorRate := float32(0)
		memoryUsage := float32(0)

		if requestCount > 0 {
			errorRate = float32(errorCount) / float32(requestCount)
		}
		if totalMemoryUsage > 0 {
			memoryUsage = float32(totalMemoryUsage) / float32(len(snapshotGroup))
		}

		result = append(result, TimeSeriesPoint{
			Timestamp:       t,
			RequestCount:    requestCount,
			ErrorRate:       errorRate,
			ResponseTimeP50: utils.GetTimePercentile(p50, 50),
			ResponseTimeP95: utils.GetTimePercentile(p95, 95),
			ResponseTimeP99: utils.GetTimePercentile(p99, 99),
			MemoryUsage:     memoryUsage,
		})
	}

	return result, nil
}

func getWeekStartAndEnd(now time.Time) (time.Time, time.Time) {
	weekday := now.Weekday()
	monday := now.AddDate(0, 0, -int(weekday)+1)
	if weekday == time.Sunday {
		monday = now.AddDate(0, 0, -6)
	}
	weekStart := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, now.Location())
	weekEnd := weekStart.AddDate(0, 0, 7)

	return weekStart, weekEnd
}
