package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "page parameter should be more than 1"})
		return
	}
	if limit < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit parameter should be more than 0"})
		return
	}

	// Get total count
	total, err := gr.leaderboardService.GetTotalEntries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch leaderboard"})
		return
	}

	// Get entries
	entries, err := gr.leaderboardService.GetLeaderboard(limit, (page-1)*limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch leaderboard"})
		return
	}

	sessionId, err := c.Cookie(sessionIdCookie)
	userId, err := c.Cookie(sessionIdCookie)

	apiEntries := make([]LeaderboardEntry, len(entries))
	for i, entry := range entries {
		apiEntries[i] = ToApiLeaderboardEntry(entry, userId, sessionId)
	}

	fmt.Println(entries, apiEntries, err)

	c.JSON(http.StatusOK, GetLeaderboardResponse{
		Entries: apiEntries,
		Total:   total,
	})
}
