package routes

import (
	"net/http"

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
	AutoVoice bool `json:"autoVoice" binding:"required"`
}

type StartGameResponse struct {
	SessionID  string `json:"sessionId"`
	CurrentWord string `json:"currentWord"`
	Token      string `json:"token"`
}

func (gr *GameRoutes) StartGame(c *gin.Context) {
	var req StartGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Create session
	session, err := gr.sessionService.Create(req.AutoVoice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	// Start game
	session, err = gr.gameService.StartGame(session.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start game"})
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
		"sessionId":      session.ID,
		"currentWord":    session.CurrentWord,
		"token":          session.CurrentWordToken,
		"remainingSeconds": session.RemainingSeconds,
	})

	// Create a ticker for sending updates every second
	ticker := make(chan struct{}, 1)
	done := make(chan bool, 1)

	// Goroutine to emit updates
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				session, err := gr.sessionService.FindByID(sessionID)
				if err != nil {
					return
				}

				// Send tick event
				c.SSEvent("tick", gin.H{
					"remainingSeconds": session.RemainingSeconds,
					"phase":            session.Phase,
					"score":            session.Score,
				})

				// If game is finished, send finish event and close
				if session.Phase == "finished" {
					c.SSEvent("finish", gin.H{
						"score":         session.Score,
						"totalAttempts": session.TotalAttempts,
						"accuracy":      float32(session.Score) / float32(session.TotalAttempts) * 100,
					})
					return
				}

				// Wait 1 second before next update
				<-ticker
			}
		}
	}()

	// Send ticker signals
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				ticker <- struct{}{}
				// Sleep is done in the client loop
			}
		}
	}()

	// Keep connection open
	c.Stream(func(w *gin.ResponseWriter) bool {
		return true
	})
	done <- true
}
