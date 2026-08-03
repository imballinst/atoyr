package services

import (
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
	db     *gorm.DB
	Now    func() time.Time
	topics []string
}

type TopicStats struct {
	Topic             string `json:"topic"`
	SessionsToday     int64  `json:"sessionsToday"`
	SessionsThisWeek  int64  `json:"sessionsThisWeek"`
	SessionsThisMonth int64  `json:"sessionsThisMonth"`
	ActiveGames       int64  `json:"activeGames"`
	TotalSessions     int64  `json:"totalSessions"`
}

type ModeStats struct {
	Mode              string       `json:"mode"`
	SessionsToday     int64        `json:"sessionsToday"`
	SessionsThisWeek  int64        `json:"sessionsThisWeek"`
	SessionsThisMonth int64        `json:"sessionsThisMonth"`
	ActiveGames       int64        `json:"activeGames"`
	TotalSessions     int64        `json:"totalSessions"`
	Topics            []TopicStats `json:"topics"`
}

type StatsResponse struct {
	TotalSessions     int64       `json:"totalSessions"`
	SessionsToday     int64       `json:"sessionsToday"`
	SessionsThisWeek  int64       `json:"sessionsThisWeek"`
	SessionsThisMonth int64       `json:"sessionsThisMonth"`
	ActiveGames       int64       `json:"activeGames"`
	ModeBreakdown     []ModeStats `json:"modeBreakdown"`
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

func NewStatsService(db *gorm.DB, topics []string) *StatsService {
	return &StatsService{
		db:     db,
		Now:    time.Now,
		topics: topics,
	}
}

// BlindSupportedTopics lists the topics that can be played in blind mode.
// When adding a new topic, decide whether it supports blind mode and update
// this set (mirrored by the game routes validation).
var BlindSupportedTopics = map[string]bool{
	"english-words": true,
}

func modeTopicKey(mode, topic string) string {
	return mode + "\x00" + topic
}

func (s *StatsService) topicsForMode(mode string) []string {
	if mode == "blind" {
		result := []string{}
		for _, topic := range s.topics {
			if BlindSupportedTopics[topic] {
				result = append(result, topic)
			}
		}
		return result
	}
	return s.topics
}

func (s *StatsService) GetStats(includeTotalSessions bool) (*StatsResponse, error) {
	now := s.Now().UTC()
	today := now.Truncate(24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour).Truncate(24 * time.Hour)
	weekStart, weekEnd := getWeekStartAndEnd(now)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonthStart := monthStart.AddDate(0, 1, 0)

	var stats StatsResponse

	// 1. Single query: all sessions this month — categorize into today/week/month in Go
	type monthlyRow struct {
		Mode      string
		Topic     string
		CreatedAt time.Time
	}
	var monthly []monthlyRow
	result := s.db.Model(&models.SessionEntity{}).
		Select("mode, topic, created_at").
		Where("created_at >= ? AND created_at < ?", monthStart, nextMonthStart).
		Find(&monthly)
	if result.Error != nil {
		return nil, result.Error
	}

	type periodCounts struct {
		today, week, month int64
	}

	allCounts := periodCounts{}
	modeTopicCounts := map[string]*periodCounts{}

	for _, row := range monthly {
		allCounts.month++
		if !row.CreatedAt.Before(today) && row.CreatedAt.Before(tomorrow) {
			allCounts.today++
		}
		if !row.CreatedAt.Before(weekStart) && row.CreatedAt.Before(weekEnd) {
			allCounts.week++
		}

		key := modeTopicKey(row.Mode, row.Topic)
		if modeTopicCounts[key] == nil {
			modeTopicCounts[key] = &periodCounts{}
		}
		mc := modeTopicCounts[key]
		mc.month++
		if !row.CreatedAt.Before(today) && row.CreatedAt.Before(tomorrow) {
			mc.today++
		}
		if !row.CreatedAt.Before(weekStart) && row.CreatedAt.Before(weekEnd) {
			mc.week++
		}
	}

	stats.SessionsToday = allCounts.today
	stats.SessionsThisWeek = allCounts.week
	stats.SessionsThisMonth = allCounts.month

	// 2. Single query: active games grouped by mode and topic
	type groupCount struct {
		Mode  string
		Topic string
		Count int64
	}
	var activeCounts []groupCount
	result = s.db.Model(&models.SessionEntity{}).
		Select("mode, topic, COUNT(*) as count").
		Where("phase != ?", "finished").
		Group("mode, topic").
		Find(&activeCounts)
	if result.Error != nil {
		return nil, result.Error
	}

	activeByModeTopic := map[string]int64{}
	totalActive := int64(0)
	for _, a := range activeCounts {
		activeByModeTopic[modeTopicKey(a.Mode, a.Topic)] = a.Count
		totalActive += a.Count
	}
	stats.ActiveGames = totalActive

	// 3. Total sessions (on-demand, grouped by mode and topic)
	var totalByModeTopic map[string]int64
	if includeTotalSessions {
		var totalCounts []groupCount
		result = s.db.Model(&models.SessionEntity{}).
			Select("mode, topic, COUNT(*) as count").
			Group("mode, topic").
			Find(&totalCounts)
		if result.Error != nil {
			return nil, result.Error
		}

		totalByModeTopic = map[string]int64{}
		totalAll := int64(0)
		for _, t := range totalCounts {
			totalByModeTopic[modeTopicKey(t.Mode, t.Topic)] = t.Count
			totalAll += t.Count
		}
		stats.TotalSessions = totalAll
	} else {
		stats.TotalSessions = -1
	}

	// Build mode breakdown with per-topic rows
	for _, mode := range []string{"vanilla", "blind"} {
		ms := ModeStats{
			Mode:   mode,
			Topics: []TopicStats{},
		}
		for _, topic := range s.topicsForMode(mode) {
			key := modeTopicKey(mode, topic)
			tc := modeTopicCounts[key]
			ts := TopicStats{Topic: topic, ActiveGames: activeByModeTopic[key]}
			if tc != nil {
				ts.SessionsToday = tc.today
				ts.SessionsThisWeek = tc.week
				ts.SessionsThisMonth = tc.month
			}
			if includeTotalSessions {
				ts.TotalSessions = totalByModeTopic[key]
			} else {
				ts.TotalSessions = -1
			}

			ms.SessionsToday += ts.SessionsToday
			ms.SessionsThisWeek += ts.SessionsThisWeek
			ms.SessionsThisMonth += ts.SessionsThisMonth
			ms.ActiveGames += ts.ActiveGames
			ms.TotalSessions += ts.TotalSessions
			ms.Topics = append(ms.Topics, ts)
		}
		if !includeTotalSessions {
			ms.TotalSessions = -1
		}
		stats.ModeBreakdown = append(stats.ModeBreakdown, ms)
	}

	return &stats, nil
}

func (s *StatsService) GetTimeSeries(period, granularity string) (*TimeSeriesResponse, error) {
	startTime, endTime, interval, effectiveGranularity, err := s.parseTimeRange(period, granularity)
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

	data, err := aggregateSnapshots(snapshots, startTime, endTime, interval)
	if err != nil {
		return nil, err
	}

	return &TimeSeriesResponse{
		Period:      period,
		Granularity: effectiveGranularity,
		Data:        data,
	}, nil
}

func (s *StatsService) parseTimeRange(period, granularity string) (time.Time, time.Time, time.Duration, string, error) {
	now := s.Now().UTC()

	var startTime time.Time

	switch period {
	case "1h":
		startTime = now.Add(-1 * time.Hour)
	case "24h":
		startTime = now.Add(-24 * time.Hour)
	case "7d":
		startTime = now.AddDate(0, 0, -7)
	case "30d":
		startTime = now.AddDate(0, 0, -30)
	default:
		return time.Time{}, time.Time{}, 0, "", fmt.Errorf("%s: %s", ErrorInvalidPeriod, period)
	}

	var interval time.Duration
	effectiveGranularity := granularity

	switch granularity {
	case "1m":
		interval = time.Minute
	case "5m":
		interval = 5 * time.Minute
	case "1h":
		interval = time.Hour
	case "1d":
		interval = 24 * time.Hour
	default:
		// Auto-select based on period
		switch period {
		case "1h":
			interval = time.Minute
			effectiveGranularity = "1m"
		case "24h":
			interval = time.Hour
			effectiveGranularity = "1h"
		case "7d", "30d":
			interval = 24 * time.Hour
			effectiveGranularity = "1d"
		}
	}

	return startTime, now, interval, effectiveGranularity, nil
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
