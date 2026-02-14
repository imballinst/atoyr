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
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := database.Automigrate(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

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

	assert.NoError(t, err)

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

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response StartGameResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.SessionID == "" {
		t.Error("Session ID is empty")
	}

	if response.ScrambledWord == "" {
		t.Error("Current word is empty")
	}

	if response.ScrambledWordDefinition == "" {
		t.Error("Current word definition is empty")
	}

	if response.Token == "" {
		t.Error("Token is empty")
	}

	if response.RemainingSeconds == 0 {
		t.Error("Remaining seconds is empty")
	}
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

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["correct"] != true {
		t.Error("Expected correct answer")
	}

	if response["score"].(float64) == 0 {
		t.Error("Expected score greater than 0")
	}
}

func TestLeaderboardRoutes_GetLeaderboard(t *testing.T) {
	router, _ := setupTestRouter(t)

	req, _ := http.NewRequest("GET", "/api/leaderboard", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["total"] == nil {
		t.Error("Total field is missing")
	}

	if response["entries"] == nil {
		t.Error("Entries field is missing")
	}
}

func TestLeaderboardRoutes_Pagination(t *testing.T) {
	router, _ := setupTestRouter(t)

	req, _ := http.NewRequest("GET", "/api/leaderboard?page=0&limit=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
