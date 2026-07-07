package api

import (
	"atoyr/server/internal/core"
	"atoyr/server/internal/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (gr *Server) GetApiV1Leaderboard(c *gin.Context, params GetApiV1LeaderboardParams) {
	page := 1
	limit := 10

	if params.Page != nil {
		page = *params.Page
	}
	if params.Limit != nil {
		limit = *params.Limit
	}

	if page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page parameter should be more than 0"})
		return
	}
	if limit < 1 || limit > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit parameter should be between 1 and 10"})
		return
	}

	// Get total count
	total, err := gr.leaderboardService.GetTotalEntries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch leaderboard"})
		utils.SendExceptionToSentry(err)
		return
	}

	// Get entries
	entries, err := gr.leaderboardService.GetLeaderboard(limit, (page-1)*limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch leaderboard"})
		utils.SendExceptionToSentry(err)
		return
	}

	sessionId, err := c.Cookie(sessionIdCookie)
	apiEntries := make([]LeaderboardEntry, len(entries))
	for i, entry := range entries {
		apiEntries[i] = ToApiLeaderboardEntry(entry, sessionId)
	}

	c.JSON(http.StatusOK, GetLeaderboardResponse{
		Entries: apiEntries,
		Total:   total,
	})
}

func (gr *Server) GetApiV1LeaderboardPercentile(c *gin.Context) {
	sessionId, err := c.Cookie(sessionIdCookie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch session ID from cookie"})
		return
	}

	session, err := gr.sessionService.FindByID(sessionId)
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

	if session.Phase != core.SessionPhaseFinished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session not finished yet"})
		return
	}

	percentile, err := gr.leaderboardService.GetPercentile(sessionId, session.Score)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch leaderboard percentile"})
		utils.SendExceptionToSentry(err)
		return
	}

	c.JSON(http.StatusOK, GetLeaderboardPercentileResponse{
		Percentile: percentile,
	})
}
