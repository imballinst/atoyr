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

	if !req.Mode.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request, mode is invalid"})
		return
	}

	if !req.Topic.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request, topic is invalid"})
		return
	}

	if req.Mode == Blind && req.Topic == IndonesianPoliticianQuotes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "blind mode is not supported for the indonesian-politician-quotes topic"})
		return
	}

	// Create session
	session, err := gr.sessionService.Create(*req.AutoVoice, req.ItemsUsed, string(req.Mode), string(req.Topic), gr.sessionOptions.Duration)
	if err != nil {
		log.Println("Failed to create session:", err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session, " + err.Error()})
		utils.SendExceptionToSentry(err)
		return
	}

	// Start game
	session, err = gr.gameService.StartGame(session.ID)
	if err != nil {
		log.Println("Failed to start game:", err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start game, " + err.Error()})
		utils.SendExceptionToSentry(err)
		return
	}

	c.SetCookie(sessionIdCookie, session.ID, 3600, "/", "", false, true)
	c.JSON(http.StatusCreated, StartGameResponse{
		SessionId:               session.ID,
		ScrambledWord:           utils.ScrambleWord(session.CurrentWord),
		ScrambledWordDefinition: session.CurrentWordDefinition,
		AutoVoice:               session.AutoVoice,
		Mode:                    SessionMode(session.Mode),
		Topic:                   SessionTopic(session.Topic),
		Lang:                    gr.wordService.GetTopicLang(session.Topic),
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
		} else {
			// Server-side error.
			utils.SendExceptionToSentry(err)
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
		Mode:                    SessionMode(session.Mode),
		Topic:                   SessionTopic(session.Topic),
		Lang:                    gr.wordService.GetTopicLang(session.Topic),
		RemainingSeconds:        core.SessionDurationManager.Get(session.ID),
		AutoVoice:               session.AutoVoice,
	})
}

func (gr *Server) PostApiV1GameSubmit(c *gin.Context) {
	sessionId, err := c.Cookie(sessionIdCookie)
	if err != nil {
		log.Printf("No %s cookie found, starting game without user association\n", sessionIdCookie)
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid session"})
		return
	}

	var req SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	result, err := gr.gameService.SubmitAnswer(sessionId, strings.ToLower(req.Answer), req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	timeGroups := make([][]time.Time, len(result.CorrectAttemptTimestamps))
	for i, tsGroup := range result.CorrectAttemptTimestamps {
		timeGroup := []time.Time{}

		for _, ts := range tsGroup {
			parseResult, err := time.Parse(time.RFC3339, ts)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			timeGroup = append(timeGroup, parseResult)
		}

		timeGroups[i] = timeGroup
	}

	c.JSON(http.StatusOK, SubmitAnswerResponse{
		Attempts:                 result.Attempts,
		Correct:                  result.Correct,
		CorrectAttemptTimestamps: timeGroups,
		RemainingSeconds:         result.DurationSeconds,
		Score:                    result.Score,
		ScrambledWord:            result.ScrambledWord,
		ScrambledWordDefinition:  result.ScrambledWordDefinition,
		Token:                    result.Token,
	})
}

func (gr *Server) GetApiV1GameSse(c *gin.Context) {
	sessionId, err := c.Cookie(sessionIdCookie)
	if err != nil {
		log.Printf("No %s cookie found, starting game without user association\n", sessionIdCookie)
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid session"})
		return
	}

	_, err = gr.sessionService.FindByID(sessionId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Session with ID %s not found\n", sessionId)
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		utils.SendExceptionToSentry(err)
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	c.Stream(func(w io.Writer) bool {
		time.Sleep(utils.ToDuration(gr.sessionOptions.Tick))

		remainingSeconds := core.SessionDurationManager.Get(sessionId)
		currentPhase := core.GetSessionPhase(remainingSeconds)

		if currentPhase == core.SessionPhasePlaying {
			c.SSEvent("tick", gin.H{
				"remainingSeconds": remainingSeconds,
			})
			return true
		}

		if currentPhase == core.SessionPhaseFinished {
			session, err := gr.sessionService.FindByID(sessionId)
			if err != nil {
				fmt.Println(fmt.Errorf("error when retrieving session after session finished: %v", err))
				return false
			}

			finishEvent := gin.H{
				"lastWordAnswer": session.CurrentWord,
			}

			if session.Topic == string(IndonesianPoliticianQuotes) {
				finishEvent["lastWordDefinition"] = resolveDefinition(session.CurrentWordDefinition, session.CurrentWord)
			}

			c.SSEvent("finish", finishEvent)
		}

		return false
	})
}

func resolveDefinition(definition string, word string) string {
	return strings.ReplaceAll(definition, "<template>", word)
}
