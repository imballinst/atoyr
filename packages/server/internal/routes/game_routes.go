package routes

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"atoyr/server/internal/services"

	"github.com/gin-gonic/gin"
)

type GameRoutes struct {
	gameService    *services.GameService
	sessionService *services.SessionService
}

func NewGameRoutes(gameService *services.GameService, sessionService *services.SessionService) *GameRoutes {
	return &GameRoutes{
		gameService:    gameService,
		sessionService: sessionService,
	}
}

func (gr *GameRoutes) Register(r *gin.Engine) {
	api := r.Group("/api")
	game := api.Group("/game")

	game.POST("/start", gr.StartGame)
	game.POST("/answer", gr.SubmitAnswer)
	game.GET("/sse/:sessionId", gr.SSE)
}

type StartGameRequest struct {
	AutoVoice *bool `json:"autoVoice" binding:"required"`
}

type StartGameResponse struct {
	SessionID   string `json:"sessionId"`
	CurrentWord string `json:"currentWord"`
	Token       string `json:"token"`
}

func (gr *GameRoutes) StartGame(c *gin.Context) {
	var req StartGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request, " + err.Error()})
		return
	}

	// Create session
	session, err := gr.sessionService.Create(*req.AutoVoice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session, " + err.Error()})
		return
	}

	// Start game
	session, err = gr.gameService.StartGame(session.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start game, " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, StartGameResponse{
		SessionID:   session.ID,
		CurrentWord: session.CurrentWord,
		Token:       session.CurrentWordToken,
	})
}

type SubmitAnswerRequest struct {
	SessionID string `json:"sessionId" binding:"required"`
	Answer    string `json:"answer" binding:"required"`
}

func (gr *GameRoutes) SubmitAnswer(c *gin.Context) {
	var req SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	result, err := gr.gameService.SubmitAnswer(req.SessionID, req.Answer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (gr *GameRoutes) SSE(c *gin.Context) {
	sessionID := c.Param("sessionId")

	session, err := gr.sessionService.FindByID(sessionID)
	if err != nil {
		log.Println("Failed to find session for SSE:", err)

		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Send initial state
	c.SSEvent("start", gin.H{
		"sessionId":        session.ID,
		"currentWord":      session.CurrentWord,
		"token":            session.CurrentWordToken,
		"remainingSeconds": session.RemainingSeconds,
	})

	c.Stream(func(w io.Writer) bool {
		fmt.Println("Streaming data...")
		time.Sleep(time.Second)

		session, err := gr.sessionService.FindByID(sessionID)
		if err != nil {
			fmt.Println("Cannot get session by ID: " + err.Error())
			return false
		}

		fmt.Println(session.RemainingSeconds, session.Phase)

		if session.Phase == "playing" {
			c.SSEvent("tick", gin.H{
				"remainingSeconds": session.RemainingSeconds,
				"phase":            session.Phase,
				"score":            session.Score,
			})
			return true
		}

		if session.Phase == "finished" {
			c.SSEvent("finish", gin.H{
				"score":         session.Score,
				"totalAttempts": session.TotalAttempts,
				"accuracy":      float32(session.Score) / float32(session.TotalAttempts) * 100,
			})
		}

		fmt.Println("false...")

		return false
	})
}
