package api

import (
	"atoyr/server/internal/services"
	"atoyr/server/internal/utils"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	sessionIdCookie = "session_id"
	userIdCookie    = "user_id"
)

func (gr *Server) PostApiV1GameStart(c *gin.Context) {
	var req StartGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request, " + err.Error()})
		return
	}

	userId, err := c.Cookie(userIdCookie)
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

	c.SetCookie(sessionIdCookie, session.ID, 3600, "/", "", false, true)
	c.JSON(http.StatusCreated, StartGameResponse{
		SessionId:               session.ID,
		ScrambledWord:           utils.ScrambleWord(session.CurrentWord),
		ScrambledWordDefinition: session.CurrentWordDefinition,
		Token:                   session.CurrentWordToken,
		RemainingSeconds:        session.RemainingSeconds,
	})
}

func (gr *Server) PostApiV1GameContinue(c *gin.Context) {
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
		if err == services.ErrSessionNotPlayingYet || err == services.ErrSessionAlreadyFinished {
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
		SessionId:               session.ID,
		ScrambledWord:           utils.ScrambleWord(session.CurrentWord),
		ScrambledWordDefinition: session.CurrentWordDefinition,
		Token:                   session.CurrentWordToken,
		RemainingSeconds:        session.RemainingSeconds,
		AutoVoice:               session.AutoVoice,
	})
}

func (gr *Server) PostApiV1GameSubmit(c *gin.Context) {
	var req SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	result, err := gr.gameService.SubmitAnswer(req.SessionId, strings.ToLower(req.Answer), req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (gr *Server) PostApiV1GameRegister(c *gin.Context) {
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

func (gr *Server) GetApiV1GameSseSessionId(c *gin.Context, sessionId string) {
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

		if session.Phase == services.SessionPhaseFinished {
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
