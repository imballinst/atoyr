package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"atoyr/server/internal/database"
	"atoyr/server/internal/middleware"
	"atoyr/server/internal/services"
	"atoyr/server/internal/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *services.SessionService) {
	// Set Gin to release mode for tests
	gin.SetMode(gin.TestMode)

	// Create test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = database.Automigrate(db)
	assert.NoError(t, err)

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

	// Create router
	router := gin.New()
	router.Use(middleware.CORSMiddleware())

	// Register routes
	gameRoutes := NewGameRoutes(gameService, sessionService, inventoryService)
	gameRoutes.Register(router)

	leaderboardRoutes := NewLeaderboardRoutes(leaderboardService)
	leaderboardRoutes.Register(router)

	return router, sessionService
}

func TestGameRoutes_StartGame(t *testing.T) {
	router, _ := setupTestRouter(t)

	autoVoice := true
	payload := StartGameRequest{AutoVoice: &autoVoice, ItemsUsed: []string{}}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/game/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response StartGameResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.NotEqual(t, "", response.SessionID)
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

	req, _ := http.NewRequest("POST", "/api/game/start", bytes.NewBuffer(startBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var startResponse StartGameResponse
	json.Unmarshal(w.Body.Bytes(), &startResponse)

	session, _ := sessionService.FindByID(startResponse.SessionID)

	// Submit answer
	answerPayload := SubmitAnswerRequest{
		SessionID: startResponse.SessionID,
		Answer:    session.CurrentWord,
	}
	answerBody, _ := json.Marshal(answerPayload)

	req, _ = http.NewRequest("POST", "/api/game/answer", bytes.NewBuffer(answerBody))
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
	router, _ := setupTestRouter(t)

	req, _ := http.NewRequest("GET", "/api/leaderboard", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.NotNil(t, response["total"])
	assert.NotNil(t, response["entries"])
}

func TestLeaderboardRoutes_Pagination(t *testing.T) {
	router, _ := setupTestRouter(t)

	req, _ := http.NewRequest("GET", "/api/leaderboard?page=0&limit=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
