package routes

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"atoyr/server/internal/services"
	"atoyr/server/internal/utils"

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

	// TODO: start should also be able to consume items if any
	game.POST("/start", gr.StartGame)
	game.POST("/answer", gr.SubmitAnswer)
	game.GET("/sse/:sessionId", gr.SSE)
}

type StartGameRequest struct {
	AutoVoice *bool `json:"autoVoice" binding:"required"`
}

type StartGameResponse struct {
	SessionID               string `json:"sessionId"`
	ScrambledWord           string `json:"scrambledWord"`
	ScrambledWordDefinition string `json:"scrambledWordDefinition"`
	Token                   string `json:"token"`
	RemainingSeconds        int32  `json:"remainingSeconds"`
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
		log.Println("Failed to create session:", err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session, " + err.Error()})
		return
	}

	// Start game
	session, err = gr.gameService.StartGame(session.ID)
	if err != nil {
		log.Println("Failed to start game:", err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start game, " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, StartGameResponse{
		SessionID:               session.ID,
		ScrambledWord:           utils.ScrambleWord(session.CurrentWord),
		ScrambledWordDefinition: session.CurrentWordDefinition,
		Token:                   session.CurrentWordToken,
		RemainingSeconds:        session.RemainingSeconds,
	})
}

type SubmitAnswerRequest struct {
	SessionID string `json:"sessionId" binding:"required"`
	Answer    string `json:"answer" binding:"required"`
	Token     string `json:"token" binding:"required"`
}

func (gr *GameRoutes) SubmitAnswer(c *gin.Context) {
	var req SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	result, err := gr.gameService.SubmitAnswer(req.SessionID, strings.ToLower(req.Answer), req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (gr *GameRoutes) SSE(c *gin.Context) {
	sessionID := c.Param("sessionId")

	_, err := gr.sessionService.FindByID(sessionID)
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

			time.Sleep(time.Second)
		}

		fmt.Println("closing SSE connection...")

		return false
	})
}
