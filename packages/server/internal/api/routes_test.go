package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"atoyr/server/internal/middleware"
	"atoyr/server/internal/services"
	"atoyr/server/internal/services/domainmodels"
	"atoyr/server/internal/testutils"
	"atoyr/server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *services.SessionService) {
	// Set Gin to release mode for tests
	gin.SetMode(gin.TestMode)

	db := testutils.SetupTestDB(t)

	// Create services
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

	_, itemInfoMap := testutils.SetupItemInfoMap([]string{"test"})

	sessionService := services.NewSessionService(db)
	leaderboardService := services.NewLeaderboardService(db)
	gameService := services.NewGameService(sessionService, wordService, leaderboardService, itemInfoMap)
	inventoryService := services.NewInventoryService(db, itemInfoMap)
	userService := services.NewUserService(sessionService, inventoryService)

	// Create router
	router := gin.New()
	router.Use(middleware.CORSMiddleware())

	// Register routes
	server := NewServer(gameService, sessionService, inventoryService, leaderboardService, userService)
	RegisterHandlers(router, server)

	return router, sessionService
}

func TestGameRoutes_StartGame(t *testing.T) {
	router, _ := setupTestRouter(t)

	autoVoice := true
	payload := StartGameRequest{AutoVoice: &autoVoice, ItemsUsed: []string{}}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/v1/game/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response StartGameResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.NotEqual(t, "", response.SessionId)
	assert.NotEqual(t, "", response.ScrambledWord)
	assert.NotEqual(t, "", response.ScrambledWordDefinition)
	assert.NotEqual(t, "", response.Token)
	assert.NotEqual(t, int32(0), response.RemainingSeconds)

	cookie, err := http.ParseSetCookie(w.Header().Get("set-cookie"))
	assert.NoError(t, err)

	assert.NotEmpty(t, cookie)
}

func TestGameRoutes_SubmitAnswer(t *testing.T) {
	router, sessionService := setupTestRouter(t)

	// Start a game first
	autoVoice := false
	startPayload := StartGameRequest{AutoVoice: &autoVoice}
	startBody, _ := json.Marshal(startPayload)

	req, _ := http.NewRequest("POST", "/api/v1/game/start", bytes.NewBuffer(startBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var startResponse StartGameResponse
	assert.Equal(t, http.StatusCreated, w.Result().StatusCode)
	json.Unmarshal(w.Body.Bytes(), &startResponse)

	session, _ := sessionService.FindByID(startResponse.SessionId)

	// Submit answer
	answerPayload := SubmitAnswerRequest{
		SessionId: startResponse.SessionId,
		Answer:    session.CurrentWord,
		Token:     session.CurrentWordToken,
	}
	answerBody, _ := json.Marshal(answerPayload)

	req, _ = http.NewRequest("POST", "/api/v1/game/submit", bytes.NewBuffer(answerBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response services.SubmitAnswerResult
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, true, response.Correct)
	assert.Equal(t, int32(1), response.Score)
}

func TestLeaderboardRoutes_GetLeaderboard(t *testing.T) {
	router, sessionService := setupTestRouter(t)

	sessionIDs := []string{}
	for range 5 {
		session, _ := sessionService.Create(false, []string{})
		sessionIDs = append(sessionIDs, session.ID)
	}

	getSessionDomainPatchInfo := func(sessionID string, score, totalAttempts int32) domainmodels.SessionDomain {
		return domainmodels.SessionDomain{
			ID:            sessionID,
			Score:         score,
			TotalAttempts: totalAttempts,
			Accuracy:      utils.CalculateAccuracy(score, totalAttempts),
		}
	}

	patchedSessionsInfo := []domainmodels.SessionDomain{
		getSessionDomainPatchInfo(sessionIDs[0], 10, 15),
		getSessionDomainPatchInfo(sessionIDs[1], 10, 12),
		getSessionDomainPatchInfo(sessionIDs[2], 10, 10),
		getSessionDomainPatchInfo(sessionIDs[3], 10, 20),
		getSessionDomainPatchInfo(sessionIDs[4], 10, 30),
	}

	for _, sessionInfo := range patchedSessionsInfo {
		session, err := sessionService.FindByID(sessionInfo.ID)
		assert.NoError(t, err)

		session.Score = sessionInfo.Score
		session.TotalAttempts = sessionInfo.TotalAttempts
		session.Accuracy = sessionInfo.Accuracy
		session.Phase = services.SessionPhaseFinished

		err = sessionService.Update(session)
		assert.NoError(t, err)
	}

	req, _ := http.NewRequest("GET", "/api/v1/leaderboard", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response GetLeaderboardResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, int64(5), response.Total)
	assert.Len(t, response.Entries, 5)

	expectedLeaderboardOrder := []domainmodels.SessionDomain{patchedSessionsInfo[2], patchedSessionsInfo[1], patchedSessionsInfo[0], patchedSessionsInfo[3], patchedSessionsInfo[4]}
	for i, session := range response.Entries {
		assert.Equal(t, expectedLeaderboardOrder[i].Score, session.Score)
		assert.Equal(t, expectedLeaderboardOrder[i].TotalAttempts, session.TotalAttempts)
		assert.Equal(t, expectedLeaderboardOrder[i].Accuracy, session.Accuracy)

		// Ensure the IDs are all masked.
		assert.Equal(t, utils.MaskSessionID(expectedLeaderboardOrder[i].ID), session.Id)
	}
}

func TestLeaderboardRoutes_Pagination(t *testing.T) {
	router, _ := setupTestRouter(t)

	req, _ := http.NewRequest("GET", "/api/v1/leaderboard?page=1&limit=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
