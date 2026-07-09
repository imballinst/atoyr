package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"atoyr/server/internal/services"
	"atoyr/server/internal/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupHealthTestRouter(t *testing.T, dbHealthy bool) (*gin.Engine, *services.HealthService) {
	gin.SetMode(gin.TestMode)

	db := testutils.SetupTestDB(t)
	if !dbHealthy {
		sqlDB, err := db.DB()
		assert.NoError(t, err)
		assert.NoError(t, sqlDB.Close())
	}

	healthService := services.NewHealthService(db, "atoyr", "test-hash")

	router := gin.New()
	RegisterHealthRoutes(router, healthService)

	return router, healthService
}

func TestHealthRoutes_GetHealth_Healthy(t *testing.T) {
	router, _ := setupHealthTestRouter(t, true)

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response services.HealthResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "atoyr", response.Name)
	assert.Equal(t, "test-hash", response.GitHash)
	assert.True(t, response.Dependencies.Database)
}

func TestHealthRoutes_GetHealth_DatabaseUnhealthy(t *testing.T) {
	router, _ := setupHealthTestRouter(t, false)

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var response services.HealthResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Dependencies.Database)
}
