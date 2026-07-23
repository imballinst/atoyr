package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"atoyr/server/internal/core"
	"atoyr/server/internal/middleware"
	"atoyr/server/internal/models"
	"atoyr/server/internal/services"
	"atoyr/server/internal/testutils"
	"atoyr/server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, *services.SessionService, *core.SessionStore) {
	return setupTestRouterWithWordDefinition(t, nil)
}

func setupTestRouterWithWordDefinition(t *testing.T, wordDefinitionsParam []services.WordDefinition) (*gin.Engine, *gorm.DB, *services.SessionService, *core.SessionStore) {
	// Set Gin to release mode for tests
	gin.SetMode(gin.TestMode)

	db := testutils.SetupTestDB(t)

	// Create services
	wordService := &services.WordService{}
	wordDefinitions := []services.WordDefinition{
		{Word: "hello", Definition: "test"},
		{Word: "world", Definition: "test"},
		{Word: "apple", Definition: "test"},
		{Word: "banana", Definition: "test"},
		{Word: "cherry", Definition: "test"},
		{Word: "dragon", Definition: "test"},
		{Word: "elephant", Definition: "test"},
		{Word: "forest", Definition: "test"},
		{Word: "guitar", Definition: "test"},
		{Word: "horizon", Definition: "test"},
	}

	if wordDefinitionsParam != nil {
		wordDefinitions = wordDefinitionsParam
	}

	wordService.SetWords(wordDefinitions)

	sessionService := services.NewSessionService(db)
	agsSyncService := &services.AGSSyncService{}
	leaderboardService := services.NewLeaderboardService(db, sessionService, agsSyncService)
	store := core.NewSessionStore()
	gameService := services.NewGameService(sessionService, store, wordService, leaderboardService, agsSyncService, testutils.TestSessionOptions)

	// Create router
	router := gin.New()
	router.Use(middleware.CORSMiddleware())

	// Register routes
	server := NewServer(gameService, sessionService, leaderboardService, testutils.TestSessionOptions)
	RegisterHandlers(router, server)

	return router, db, sessionService, store
}

func setupAdminTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, *services.SessionService, *services.StatsService) {
	gin.SetMode(gin.TestMode)

	db := testutils.SetupTestDB(t)

	wordService := &services.WordService{}
	wordService.SetWords([]services.WordDefinition{
		{Word: "hello", Definition: "test"},
		{Word: "world", Definition: "test"},
		{Word: "apple", Definition: "test"},
		{Word: "banana", Definition: "test"},
		{Word: "cherry", Definition: "test"},
		{Word: "dragon", Definition: "test"},
		{Word: "elephant", Definition: "test"},
		{Word: "forest", Definition: "test"},
		{Word: "guitar", Definition: "test"},
		{Word: "horizon", Definition: "test"},
	})

	sessionService := services.NewSessionService(db)
	agsSyncService := &services.AGSSyncService{}
	leaderboardService := services.NewLeaderboardService(db, sessionService, agsSyncService)
	gameService := services.NewGameService(sessionService, core.NewSessionStore(), wordService, leaderboardService, agsSyncService, testutils.TestSessionOptions)
	metricsCollector := middleware.NewMetricsCollector()
	statsService := services.NewStatsService(db)

	statsService.Now = testutils.NowMockFn

	router := gin.New()
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.Metrics(metricsCollector))

	server := NewServer(gameService, sessionService, leaderboardService, testutils.TestSessionOptions)
	RegisterHandlers(router, server)
	RegisterAdminRoutes(router, statsService, middleware.NewNoopAuthMiddleware())

	return router, db, sessionService, statsService
}

func finishSessionForLeaderboard(db *gorm.DB, sessionID string, score, totalAttempts int32) {
	db.Model(&models.LeaderboardSessionEntity{ID: sessionID}).Updates(map[string]any{
		"score":          score,
		"total_attempts": totalAttempts,
		"accuracy":       utils.CalculateAccuracy(score, totalAttempts),
		"phase":          core.SessionPhaseFinished,
		"ends_at":        time.Now(),
	})
}

func TestGameRoutes_StartGame(t *testing.T) {
	router, _, _, _ := setupTestRouter(t)

	autoVoice := false
	payload := StartGameRequest{Mode: Vanilla, AutoVoice: &autoVoice, ItemsUsed: []string{}}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/v1/game/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response StartGameResponse
	json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.NotEqual(t, "", response.SessionId)
	assert.NotEqual(t, "", response.ScrambledWord)
	assert.NotEqual(t, "", response.ScrambledWordDefinition)
	assert.NotEqual(t, "", response.Token)
	assert.Equal(t, int32(1), response.RemainingSeconds) // 1 because of no autoVoice bonus.

	cookie, err := http.ParseSetCookie(recorder.Header().Get("set-cookie"))
	assert.NoError(t, err)
	assert.NotEmpty(t, cookie)

	// Get SSE, it should return no error.
	req, _ = http.NewRequest("GET", "/api/v1/game/sse", nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: sessionIdCookie, Value: response.SessionId})

	recorder = testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGameRoutes_StartGame_WithAutoVoice(t *testing.T) {
	router, _, _, _ := setupTestRouter(t)

	autoVoice := true
	payload := StartGameRequest{Mode: Vanilla, AutoVoice: &autoVoice, ItemsUsed: []string{}}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/v1/game/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response StartGameResponse
	json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.NotEqual(t, "", response.SessionId)
	assert.NotEqual(t, "", response.ScrambledWord)
	assert.NotEqual(t, "", response.ScrambledWordDefinition)
	assert.NotEqual(t, "", response.Token)
	assert.Equal(t, int32(6), response.RemainingSeconds) // 1 + 5, because of auto voice bonus.

	cookie, err := http.ParseSetCookie(recorder.Header().Get("set-cookie"))
	assert.NoError(t, err)
	assert.NotEmpty(t, cookie)

	// Get SSE, it should return no error.
	req, _ = http.NewRequest("GET", "/api/v1/game/sse", nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: sessionIdCookie, Value: response.SessionId})

	recorder = testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGameRoutes_StartGame_WithInvalidMode(t *testing.T) {
	router, _, _, _ := setupTestRouter(t)

	autoVoice := true
	payload := StartGameRequest{Mode: "randommode", AutoVoice: &autoVoice, ItemsUsed: []string{}}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/v1/game/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	cookie, err := http.ParseSetCookie(recorder.Header().Get("set-cookie"))
	assert.Error(t, err)
	assert.Empty(t, cookie)
}

func TestGameRoutes_StartGame_GetInvalidSSESession(t *testing.T) {
	router, _, _, _ := setupTestRouter(t)

	autoVoice := true
	payload := StartGameRequest{Mode: Vanilla, AutoVoice: &autoVoice, ItemsUsed: []string{}}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/v1/game/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response StartGameResponse
	json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.NotEqual(t, "", response.SessionId)
	assert.NotEqual(t, "", response.ScrambledWord)
	assert.NotEqual(t, "", response.ScrambledWordDefinition)
	assert.NotEqual(t, "", response.Token)
	assert.NotEqual(t, int32(0), response.RemainingSeconds)

	cookie, err := http.ParseSetCookie(recorder.Header().Get("set-cookie"))
	assert.NoError(t, err)
	assert.NotEmpty(t, cookie)

	// Get SSE, it should return no error.
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/game/sse/%s", "invalid-session"), nil)
	req.Header.Set("Content-Type", "application/json")

	recorder = testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestGameRoutes_FinishGame(t *testing.T) {
	router, _, sessionService, _ := setupTestRouterWithWordDefinition(t, []services.WordDefinition{
		{
			Word:       "apple",
			Definition: "A fruit with red color",
		},
	})

	autoVoice := true
	payload := StartGameRequest{Mode: Vanilla, AutoVoice: &autoVoice, ItemsUsed: []string{}}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/v1/game/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response StartGameResponse
	json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.NotEqual(t, "", response.SessionId)
	assert.NotEqual(t, "", response.ScrambledWord)
	assert.NotEqual(t, "", response.ScrambledWordDefinition)
	assert.NotEqual(t, "", response.Token)
	assert.NotEqual(t, int32(0), response.RemainingSeconds)

	cookie, err := http.ParseSetCookie(recorder.Header().Get("set-cookie"))
	assert.NoError(t, err)
	assert.NotEmpty(t, cookie)

	// Get SSE, it should return no error.
	sseReq, _ := http.NewRequest("GET", "/api/v1/game/sse", nil)
	sseReq.Header.Set("Content-Type", "application/json")
	sseReq.AddCookie(&http.Cookie{Name: sessionIdCookie, Value: response.SessionId})

	sseRecorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(sseRecorder, sseReq)

	assert.Equal(t, http.StatusOK, sseRecorder.Code)

	// End the session so we can get the finished stream event.
	session, err := sessionService.FindByID(response.SessionId)
	assert.NoError(t, err)

	session.Phase = core.SessionPhaseFinished

	err = sessionService.Update(session)
	assert.NoError(t, err)

	// Sleep, then check the event message sent.
	time.Sleep(1500 * time.Millisecond)

	var lastTick map[string]any
	json.Unmarshal(sseRecorder.Body.Bytes(), &lastTick)
}

func TestGameRoutes_SubmitAnswer(t *testing.T) {
	router, _, _, store := setupTestRouter(t)

	// Start a game first
	autoVoice := false
	startPayload := StartGameRequest{Mode: Vanilla, AutoVoice: &autoVoice}
	startBody, _ := json.Marshal(startPayload)

	req, _ := http.NewRequest("POST", "/api/v1/game/start", bytes.NewBuffer(startBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	var startResponse StartGameResponse
	assert.Equal(t, http.StatusCreated, recorder.Result().StatusCode)
	json.Unmarshal(recorder.Body.Bytes(), &startResponse)

	state := store.Get(startResponse.SessionId)

	// Submit answer
	answerPayload := SubmitAnswerRequest{
		Answer: state.CurrentWord,
		Token:  state.CurrentWordToken,
	}
	answerBody, _ := json.Marshal(answerPayload)

	req, _ = http.NewRequest("POST", "/api/v1/game/submit", bytes.NewBuffer(answerBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: sessionIdCookie, Value: state.ID})

	recorder = testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response services.SubmitAnswerResult
	json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Equal(t, true, response.Correct)
	assert.Equal(t, int32(1), response.Score)
}

func TestLeaderboardRoutes_GetLeaderboard(t *testing.T) {
	router, db, sessionService, _ := setupTestRouter(t)

	sessionIDs := []string{}
	for range 5 {
		session, _ := sessionService.Create(false, []string{}, string(Vanilla), testutils.TestSessionOptions.Duration)
		sessionIDs = append(sessionIDs, session.ID)
	}

	type patchInfo struct {
		id            string
		score         int32
		totalAttempts int32
	}

	getPatchInfo := func(sessionID string, score, totalAttempts int32) patchInfo {
		return patchInfo{sessionID, score, totalAttempts}
	}

	patchedSessionsInfo := []patchInfo{
		getPatchInfo(sessionIDs[0], 10, 15),
		getPatchInfo(sessionIDs[1], 10, 12),
		getPatchInfo(sessionIDs[2], 10, 10),
		getPatchInfo(sessionIDs[3], 10, 20),
		getPatchInfo(sessionIDs[4], 10, 30),
	}

	for _, sessionInfo := range patchedSessionsInfo {
		finishSessionForLeaderboard(db, sessionInfo.id, sessionInfo.score, sessionInfo.totalAttempts)
	}

	req, _ := http.NewRequest("GET", "/api/v1/leaderboard", nil)
	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response GetLeaderboardResponse
	json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Equal(t, int64(5), response.Total)
	assert.Len(t, response.Entries, 5)

	expectedLeaderboardOrder := []patchInfo{patchedSessionsInfo[2], patchedSessionsInfo[1], patchedSessionsInfo[0], patchedSessionsInfo[3], patchedSessionsInfo[4]}
	for i, session := range response.Entries {
		assert.Equal(t, expectedLeaderboardOrder[i].score, session.Score)
		assert.Equal(t, expectedLeaderboardOrder[i].totalAttempts, session.TotalAttempts)
		assert.Equal(t, utils.ToPercentage(utils.CalculateAccuracy(expectedLeaderboardOrder[i].score, expectedLeaderboardOrder[i].totalAttempts)), session.Accuracy)

		// Ensure the IDs are all masked.
		assert.Equal(t, utils.MaskSessionID(expectedLeaderboardOrder[i].id), session.Id)
	}
}

func TestLeaderboardRoutes_GetLeaderboard_DifferentModes(t *testing.T) {
	router, db, sessionService, _ := setupTestRouter(t)

	for _, mode := range []string{string(Vanilla), string(Blind)} {
		sessionIDs := []string{}
		for range 5 {
			session, _ := sessionService.Create(false, []string{}, mode, testutils.TestSessionOptions.Duration)
			sessionIDs = append(sessionIDs, session.ID)
		}

		type patchInfo struct {
			id            string
			score         int32
			totalAttempts int32
		}

		getPatchInfo := func(sessionID string, score, totalAttempts int32) patchInfo {
			return patchInfo{sessionID, score, totalAttempts}
		}

		patchedSessionsInfo := []patchInfo{
			getPatchInfo(sessionIDs[0], 10, 15),
			getPatchInfo(sessionIDs[1], 10, 12),
			getPatchInfo(sessionIDs[2], 10, 10),
			getPatchInfo(sessionIDs[3], 10, 20),
			getPatchInfo(sessionIDs[4], 10, 30),
		}

		for _, sessionInfo := range patchedSessionsInfo {
			finishSessionForLeaderboard(db, sessionInfo.id, sessionInfo.score, sessionInfo.totalAttempts)
		}

		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/leaderboard?mode=%s", mode), nil)

		recorder := testutils.CreateTestResponseRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response GetLeaderboardResponse
		json.Unmarshal(recorder.Body.Bytes(), &response)

		assert.Equal(t, int64(5), response.Total)
		assert.Len(t, response.Entries, 5)

		expectedLeaderboardOrder := []patchInfo{patchedSessionsInfo[2], patchedSessionsInfo[1], patchedSessionsInfo[0], patchedSessionsInfo[3], patchedSessionsInfo[4]}
		for i, session := range response.Entries {
			assert.Equal(t, expectedLeaderboardOrder[i].score, session.Score)
			assert.Equal(t, expectedLeaderboardOrder[i].totalAttempts, session.TotalAttempts)
			assert.Equal(t, utils.ToPercentage(utils.CalculateAccuracy(expectedLeaderboardOrder[i].score, expectedLeaderboardOrder[i].totalAttempts)), session.Accuracy)

			// Ensure the IDs are all masked.
			assert.Equal(t, utils.MaskSessionID(expectedLeaderboardOrder[i].id), session.Id)
		}
	}
}

func TestLeaderboardRoutes_Pagination(t *testing.T) {
	router, _, _, _ := setupTestRouter(t)

	req, _ := http.NewRequest("GET", "/api/v1/leaderboard?page=1&limit=5", nil)
	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestAdminRoutes_GetStats(t *testing.T) {
	router, _, sessionService, _ := setupAdminTestRouter(t)

	// Create sessions
	session, err := sessionService.Create(false, []string{}, string(Vanilla), testutils.TestSessionOptions.Duration)
	assert.NoError(t, err)
	// Update the session to happen today.
	session.CreatedAt = time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 1, 0, time.UTC)
	session.EndsAt = time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 31, 0, time.UTC)
	sessionService.Update(session)

	session, err = sessionService.Create(true, []string{}, string(Vanilla), testutils.TestSessionOptions.Duration)
	assert.NoError(t, err)
	// Update the session to happen for some times this month (but not today and not this week).
	session.CreatedAt = time.Date(time.Now().Year(), time.Now().Month(), 20, 0, 0, 0, 0, time.UTC)
	session.EndsAt = time.Date(time.Now().Year(), time.Now().Month(), 20, 0, 0, 30, 0, time.UTC)
	sessionService.Update(session)

	req, _ := http.NewRequest("GET", "/api/v1/admin/stats", nil)
	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var stats services.StatsResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &stats)
	assert.NoError(t, err)

	assert.Equal(t, int64(1), stats.SessionsToday)
	assert.Equal(t, int64(1), stats.SessionsThisWeek)
	assert.Equal(t, int64(2), stats.SessionsThisMonth)
}

func TestAdminRoutes_GetStats_Empty(t *testing.T) {
	router, _, _, _ := setupAdminTestRouter(t)

	req, _ := http.NewRequest("GET", "/api/v1/admin/stats", nil)
	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var stats services.StatsResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &stats)
	assert.NoError(t, err)

	assert.Equal(t, int64(0), stats.SessionsToday)
	assert.Equal(t, int64(0), stats.SessionsThisWeek)
	assert.Equal(t, int64(0), stats.SessionsThisMonth)
}

func TestAdminRoutes_GetTimeseries(t *testing.T) {
	router, db, _, statsService := setupAdminTestRouter(t)

	now := statsService.Now()
	snapshot := models.MetricSnapshot{
		Timestamp:     now.Add(-5 * time.Minute),
		RequestCount:  100,
		ErrorCount4xx: 2,
		ErrorCount5xx: 1,
		ActiveGames:   3,
		MemoryUsageMb: 50,
	}
	err := db.Create(&snapshot).Error
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/v1/admin/timeseries?period=1h&granularity=5m", nil)
	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var result services.TimeSeriesResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, "1h", result.Period)
	assert.Equal(t, "5m", result.Granularity)
	assert.Len(t, result.Data, 12)
	assert.Equal(t, int64(100), result.Data[11].RequestCount)
}

func TestAdminRoutes_GetTimeseries_InvalidPeriod(t *testing.T) {
	router, _, _, _ := setupAdminTestRouter(t)

	req, _ := http.NewRequest("GET", "/api/v1/admin/timeseries?period=invalid", nil)
	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var errorResponse map[string]string
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse["error"], "invalid period")
}

func TestAdminRoutes_GetTimeseries_NoParams(t *testing.T) {
	router, _, _, _ := setupAdminTestRouter(t)

	req, _ := http.NewRequest("GET", "/api/v1/admin/timeseries", nil)
	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var result services.TimeSeriesResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, "24h", result.Period)
}
