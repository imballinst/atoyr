package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"atoyr/server/internal/middleware"
	"atoyr/server/internal/services"
	"atoyr/server/internal/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *services.SessionService) {
	return setupTestRouterWithWordDefinition(t, nil)
}

func setupTestRouterWithWordDefinition(t *testing.T, wordDefinitionsParam []services.WordDefinition) (*gin.Engine, *services.SessionService) {
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
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/game/sse/%s", response.SessionId), nil)
	req.Header.Set("Content-Type", "application/json")

	recorder = testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGameRoutes_StartGame_GetInvalidSSESession(t *testing.T) {
	router, _ := setupTestRouter(t)

	autoVoice := true
	payload := StartGameRequest{AutoVoice: &autoVoice, ItemsUsed: []string{}}
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
	router, sessionService := setupTestRouterWithWordDefinition(t, []services.WordDefinition{
		{
			Word:       "apple",
			Definition: "A fruit with red color",
		},
	})

	autoVoice := true
	payload := StartGameRequest{AutoVoice: &autoVoice, ItemsUsed: []string{}}
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
	sseReq, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/game/sse/%s", response.SessionId), nil)
	sseReq.Header.Set("Content-Type", "application/json")

	sseRecorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(sseRecorder, sseReq)

	assert.Equal(t, http.StatusOK, sseRecorder.Code)

	// End the session so we can get the finished stream event.
	session, err := sessionService.FindByID(response.SessionId)
	assert.NoError(t, err)

	session.Phase = services.SessionPhaseFinished

	err = sessionService.Update(session)
	assert.NoError(t, err)

	// Sleep, then check the event message sent.
	time.Sleep(1500 * time.Millisecond)

	var lastTick map[string]any
	json.Unmarshal(sseRecorder.Body.Bytes(), &lastTick)

	fmt.Println(lastTick)
}

func TestGameRoutes_SubmitAnswer(t *testing.T) {
	router, sessionService := setupTestRouter(t)

	// Start a game first
	autoVoice := false
	startPayload := StartGameRequest{AutoVoice: &autoVoice}
	startBody, _ := json.Marshal(startPayload)

	req, _ := http.NewRequest("POST", "/api/v1/game/start", bytes.NewBuffer(startBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	var startResponse StartGameResponse
	assert.Equal(t, http.StatusCreated, recorder.Result().StatusCode)
	json.Unmarshal(recorder.Body.Bytes(), &startResponse)

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

	recorder = testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response services.SubmitAnswerResult
	json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Equal(t, true, response.Correct)
	assert.Equal(t, int32(1), response.Score)
}

func TestLeaderboardRoutes_GetLeaderboard(t *testing.T) {
	router, _ := setupTestRouter(t)

	req, _ := http.NewRequest("GET", "/api/v1/leaderboard", nil)
	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.NotNil(t, response["total"])
	assert.NotNil(t, response["entries"])
}

func TestLeaderboardRoutes_Pagination(t *testing.T) {
	router, _ := setupTestRouter(t)

	req, _ := http.NewRequest("GET", "/api/v1/leaderboard?page=1&limit=5", nil)
	recorder := testutils.CreateTestResponseRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}
