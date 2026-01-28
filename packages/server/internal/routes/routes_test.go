package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"atoyr/server/internal/middleware"
	"atoyr/server/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"atoyr/server/internal/models"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	// Set Gin to release mode for tests
	gin.SetMode(gin.TestMode)

	// Create test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.AutoMigrate(&models.SessionEntity{}, &models.ResultEntity{}); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create services
	wordService := &services.WordService{}
	wordService.SetWords([]services.WordDefinition{
		{Word: "hello"},
		{Word: "world"},
		{Word: "apple"},
		{Word: "banana"},
		{Word: "cherry"},
		{Word: "dragon"},
		{Word: "elephant"},
		{Word: "forest"},
		{Word: "guitar"},
		{Word: "horizon"},
	})

	sessionService := services.NewSessionService(db)
	leaderboardService := services.NewLeaderboardService(db)
	gameService := services.NewGameService(sessionService, wordService, leaderboardService)

	// Create router
	router := gin.New()
	router.Use(middleware.CORSMiddleware())

	// Register routes
	gameRoutes := NewGameRoutes(gameService, sessionService)
	gameRoutes.Register(router)

	leaderboardRoutes := NewLeaderboardRoutes(leaderboardService)
	leaderboardRoutes.Register(router)

	return router
}

func TestGameRoutes_StartGame(t *testing.T) {
	router := setupTestRouter(t)

	autoVoice := true
	payload := StartGameRequest{AutoVoice: &autoVoice}
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

	if response.CurrentWord == "" {
		t.Error("Current word is empty")
	}

	if response.Token == "" {
		t.Error("Token is empty")
	}
}

func TestGameRoutes_SubmitAnswer(t *testing.T) {
	router := setupTestRouter(t)

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

	// Submit answer
	answerPayload := SubmitAnswerRequest{
		SessionID: startResponse.SessionID,
		Answer:    startResponse.CurrentWord,
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
	router := setupTestRouter(t)

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
	router := setupTestRouter(t)

	req, _ := http.NewRequest("GET", "/api/leaderboard?page=0&limit=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
