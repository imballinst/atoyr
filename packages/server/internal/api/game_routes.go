package api

import (
	"atoyr/server/internal/core"
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
)

func (gr *Server) PostApiV1GameStart(c *gin.Context) {
	var req StartGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request, " + err.Error()})
		return
	}

	// Create session
	session, err := gr.sessionService.Create(*req.AutoVoice, req.ItemsUsed, gr.sessionOptions.Duration)
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
		RemainingSeconds:        session.DurationSeconds,
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
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
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
		RemainingSeconds:        core.SessionDurationManager.Get(session.ID),
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

func (gr *Server) GetApiV1GameSseSessionId(c *gin.Context, sessionId string) {
	sessionID := c.Param("sessionId")

	_, err := gr.sessionService.FindByID(sessionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Session with ID %s not found\n", sessionId)
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	c.Stream(func(w io.Writer) bool {
		// fmt.Println("Streaming data...")
		time.Sleep(utils.ToDuration(gr.sessionOptions.Tick))

		remainingSeconds := core.SessionDurationManager.Get(sessionID)
		currentPhase := core.GetSessionPhase(remainingSeconds)

		if currentPhase == core.SessionPhasePlaying {
			c.SSEvent("tick", gin.H{
				"remainingSeconds": remainingSeconds,
			})
			return true
		}

		if currentPhase == core.SessionPhaseFinished {
			session, err := gr.sessionService.FindByID(sessionID)
			if err != nil {
				fmt.Println(fmt.Errorf("error when retrieving session after session finished: %v", err))
				return false
			}

			c.SSEvent("finish", gin.H{
				"lastWordAnswer": session.CurrentWord,
			})
		}

		fmt.Println("closing SSE connection...")

		return false
	})
}
