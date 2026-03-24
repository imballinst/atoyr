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
	"gorm.io/gorm"
)

const (
	sessionIdCookie = "session_id"
)

type GameRoutes struct {
	gameService      *services.GameService
	sessionService   *services.SessionService
	inventoryService *services.InventoryService
	userService      *services.UserService
}

func NewGameRoutes(
	gameService *services.GameService,
	sessionService *services.SessionService,
	inventoryService *services.InventoryService,
	userService *services.UserService,
) *GameRoutes {
	return &GameRoutes{
		gameService:      gameService,
		sessionService:   sessionService,
		inventoryService: inventoryService,
		userService:      userService,
	}
}

func (gr *GameRoutes) Register(r *gin.Engine) {
	api := r.Group("/api")
	game := api.Group("/game")

	// TODO: start should also be able to consume items if any
	game.POST("/start", gr.StartGame)
	game.POST("/continue", gr.ContinueGame)
	game.POST("/answer", gr.SubmitAnswer)
	game.POST("/register", gr.RegisterUser)
	game.GET("/sse/:sessionId", gr.SSE)
}

type StartGameRequest struct {
	AutoVoice *bool    `json:"autoVoice" binding:"required"`
	ItemsUsed []string `json:"itemsUsed"`
}

type StartGameResponse struct {
	SessionID               string `json:"sessionId"`
	ScrambledWord           string `json:"scrambledWord"`
	ScrambledWordDefinition string `json:"scrambledWordDefinition"`
	Token                   string `json:"token"`
	RemainingSeconds        int32  `json:"remainingSeconds"`
	AutoVoice               bool   `json:"autoVoice"`
}

func (gr *GameRoutes) StartGame(c *gin.Context) {
	var req StartGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request, " + err.Error()})
		return
	}

	userId, err := c.Cookie("user_id")
	if err != nil {
		log.Println("No user_id cookie found, starting game without user association")
	}

	err = gr.inventoryService.UseItems(req.ItemsUsed, userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error when using items, " + err.Error()})
		return
	}

	// Create session
	session, err := gr.sessionService.Create(*req.AutoVoice, req.ItemsUsed)
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

	c.SetCookie(sessionIdCookie, session.ID, int(session.RemainingSeconds), "/", "", false, true)
	c.JSON(http.StatusCreated, StartGameResponse{
		SessionID:               session.ID,
		ScrambledWord:           utils.ScrambleWord(session.CurrentWord),
		ScrambledWordDefinition: session.CurrentWordDefinition,
		Token:                   session.CurrentWordToken,
		RemainingSeconds:        session.RemainingSeconds,
	})
}

func (gr *GameRoutes) ContinueGame(c *gin.Context) {
	sessionId, err := c.Cookie(sessionIdCookie)
	if err != nil {
		log.Printf("No %s cookie found, starting game without user association\n", sessionIdCookie)
	}

	session, err := gr.gameService.ContinueGame(sessionId)
	if err == gorm.ErrRecordNotFound {
		log.Printf("Session with ID %s not found\n", sessionId)
		c.JSON(http.StatusBadRequest, gin.H{"error": "session not found"})
		return
	}
	if err != nil {
		log.Println("Failed to continue game:", err)

		status := http.StatusInternalServerError
		if err == services.ErrSessionNotPlayingYet {
			status = http.StatusBadRequest
		}

		c.JSON(status, gin.H{"error": "failed to continue game, " + err.Error()})
		return
	}
	if session == nil {
		log.Println("Failed to continue game because session does not exist")

		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to continue game, session does not exist"})
		return
	}

	c.JSON(http.StatusCreated, StartGameResponse{
		SessionID:               session.ID,
		ScrambledWord:           utils.ScrambleWord(session.CurrentWord),
		ScrambledWordDefinition: session.CurrentWordDefinition,
		Token:                   session.CurrentWordToken,
		RemainingSeconds:        session.RemainingSeconds,
		AutoVoice:               session.AutoVoice,
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

type RegisterUserRequest struct {
	Username string `json:"username" binding:"required"`
}

func (gr *GameRoutes) RegisterUser(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	sessionId, err := c.Cookie(sessionIdCookie)
	if err != nil {
		if err == http.ErrNoCookie {
			log.Printf("No %s cookie found, starting game without user association\n", sessionIdCookie)
			c.JSON(http.StatusOK, gin.H{"message": "user registered without session association"})
			return
		}

		log.Printf("Error retrieving %s from cookie, %s\n", sessionIdCookie, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve session cookie"})
		return
	}

	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username cannot be empty"})
		return
	}

	result, err := gr.userService.UpsertUserFromSession(req.Username, sessionId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (gr *GameRoutes) SSE(c *gin.Context) {
	sessionID := c.Param("sessionId")

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

		if session.Phase == services.SessionPhasePlaying {
			c.SSEvent("tick", gin.H{
				"remainingSeconds": session.RemainingSeconds,
				"phase":            session.Phase,
				"score":            session.Score,
			})
			return true
		}

		if session.Phase == "finished" {
			c.SSEvent("finish", gin.H{
				"score":                    session.Score,
				"totalAttempts":            session.TotalAttempts,
				"correctAttemptTimestamps": session.CorrectAttemptTimestamps,
			})

			time.Sleep(time.Second)
		}

		fmt.Println("closing SSE connection...")

		return false
	})
}
