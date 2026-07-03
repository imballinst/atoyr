package middleware

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"atoyr/server/internal/models"
	"atoyr/server/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MetricsCollector stores in-memory metrics
type MetricsCollector struct {
	RequestCount  atomic.Int64
	ErrorCount4xx atomic.Int64
	ErrorCount5xx atomic.Int64
	ActiveGames   atomic.Int64
	ResponseTimes []time.Duration // ring buffer
	mu            sync.Mutex
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		ResponseTimes: make([]time.Duration, 0, 1000),
	}
}

func Metrics(collector *MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		collector.RequestCount.Add(1)

		duration := time.Since(start)
		collector.addResponseTime(duration)

		status := c.Writer.Status()
		if status >= 400 && status < 500 {
			collector.ErrorCount4xx.Add(1)
		} else if status >= 500 {
			collector.ErrorCount5xx.Add(1)
		}
	}
}

func (m *MetricsCollector) StartSnapshotWorker(db *gorm.DB, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		var lastRequestCount int64
		var lastErrorCount4xx int64
		var lastErrorCount5xx int64

		for range ticker.C {
			// Calculate deltas
			currentRequests := m.RequestCount.Load()
			currentErrors4xx := m.ErrorCount4xx.Load()
			currentErrors5xx := m.ErrorCount5xx.Load()

			snapshot := models.MetricSnapshot{
				Timestamp:       time.Now().UTC(),
				RequestCount:    currentRequests - lastRequestCount,
				ErrorCount4xx:   currentErrors4xx - lastErrorCount4xx,
				ErrorCount5xx:   currentErrors5xx - lastErrorCount5xx,
				ActiveGames:     m.getActiveGames(),
				ResponseTimeP50: m.getPercentile(50),
				ResponseTimeP95: m.getPercentile(95),
				ResponseTimeP99: m.getPercentile(99),
				MemoryUsageMb:   utils.GetMemoryUsage(),
			}

			err := db.Create(&snapshot).Error
			if err != nil {
				fmt.Printf("error when saving snapshot metric: %+v\n", err)
			}

			// Update last counts
			lastRequestCount = currentRequests
			lastErrorCount4xx = currentErrors4xx
			lastErrorCount5xx = currentErrors5xx
		}
	}()
}

func (m *MetricsCollector) CleanupOldSnapshots(db *gorm.DB, retention time.Duration) {
	cutoff := time.Now().UTC().Add(-retention)
	result := db.Where("timestamp < ?", cutoff).Delete(&models.MetricSnapshot{})
	if result.Error != nil {
		fmt.Printf("error when cleaning up old metric snapshots: %+v\n", result.Error)
	} else if result.RowsAffected > 0 {
		fmt.Printf("cleaned up %d metric snapshots older than %s\n", result.RowsAffected, retention)
	}
}

func (m *MetricsCollector) addResponseTime(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Fixed size ring buffer
	if len(m.ResponseTimes) >= 1000 {
		m.ResponseTimes = m.ResponseTimes[1:]
	}
	m.ResponseTimes = append(m.ResponseTimes, d)
}

func (m *MetricsCollector) getPercentile(percentile int) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	return utils.GetTimePercentile(m.ResponseTimes, percentile)
}

func (m *MetricsCollector) getActiveGames() int64 {
	return m.ActiveGames.Load()
}
