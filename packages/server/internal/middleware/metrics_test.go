package middleware

import (
	"testing"
	"time"

	"atoyr/server/internal/models"
	"atoyr/server/internal/testutils"

	"github.com/stretchr/testify/assert"
)

func TestCleanupOldSnapshots(t *testing.T) {
	db := testutils.SetupTestDB(t)
	retention := 3 * 30 * 24 * time.Hour // 3 months
	now := time.Now().UTC()

	oldSnapshots := []models.MetricSnapshot{
		{Timestamp: now.Add(-retention - 1*time.Hour), RequestCount: 10},
		{Timestamp: now.Add(-retention - 24*time.Hour), RequestCount: 20},
		{Timestamp: now.Add(-retention - 7*24*time.Hour), RequestCount: 30},
	}
	recentSnapshots := []models.MetricSnapshot{
		{Timestamp: now.Add(-1 * time.Hour), RequestCount: 100},
		{Timestamp: now.Add(-24 * time.Hour), RequestCount: 200},
		{Timestamp: now.Add(-7 * 24 * time.Hour), RequestCount: 300},
	}

	for _, s := range append(oldSnapshots, recentSnapshots...) {
		err := db.Create(&s).Error
		assert.NoError(t, err)
	}

	var countBefore int64
	db.Model(&models.MetricSnapshot{}).Count(&countBefore)
	assert.Equal(t, int64(6), countBefore)

	collector := NewMetricsCollector()
	collector.CleanupOldSnapshots(db, retention)

	var countAfter int64
	db.Model(&models.MetricSnapshot{}).Count(&countAfter)
	assert.Equal(t, int64(3), countAfter)

	var remaining []models.MetricSnapshot
	db.Find(&remaining)
	cutoff := now.Add(-retention)
	for _, s := range remaining {
		assert.False(t, s.Timestamp.Before(cutoff), "found snapshot older than retention: %v", s.Timestamp)
	}
}

func TestCleanupOldSnapshots_NothingToDelete(t *testing.T) {
	db := testutils.SetupTestDB(t)
	retention := 3 * 30 * 24 * time.Hour
	now := time.Now().UTC()

	recentSnapshots := []models.MetricSnapshot{
		{Timestamp: now.Add(-1 * time.Hour), RequestCount: 100},
		{Timestamp: now.Add(-24 * time.Hour), RequestCount: 200},
	}
	for _, s := range recentSnapshots {
		err := db.Create(&s).Error
		assert.NoError(t, err)
	}

	collector := NewMetricsCollector()
	collector.CleanupOldSnapshots(db, retention)

	var count int64
	db.Model(&models.MetricSnapshot{}).Count(&count)
	assert.Equal(t, int64(2), count)
}
