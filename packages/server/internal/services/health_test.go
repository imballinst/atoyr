package services

import (
	"testing"

	"atoyr/server/internal/testutils"

	"github.com/stretchr/testify/assert"
)

func TestHealthService_GetHealth_DatabaseHealthy(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewHealthService(db, "atoyr", "abc123")

	health := service.GetHealth()

	assert.Equal(t, "atoyr", health.Name)
	assert.Equal(t, "abc123", health.GitHash)
	assert.True(t, health.Dependencies.Database)
	assert.False(t, health.CachedAt.IsZero())
}

func TestHealthService_GetHealth_DatabaseUnhealthy(t *testing.T) {
	db := testutils.SetupTestDB(t)

	sqlDB, err := db.DB()
	assert.NoError(t, err)
	assert.NoError(t, sqlDB.Close())

	service := NewHealthService(db, "atoyr", "abc123")

	health := service.GetHealth()

	assert.False(t, health.Dependencies.Database)
}
