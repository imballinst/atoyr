package routes

import (
	"net/http"
	"strconv"

	"atoyr/server/internal/services"
	"github.com/gin-gonic/gin"
)

type LeaderboardRoutes struct {
	leaderboardService *services.LeaderboardService
}

func NewLeaderboardRoutes(leaderboardService *services.LeaderboardService) *LeaderboardRoutes {
	return &LeaderboardRoutes{
		leaderboardService: leaderboardService,
	}
}

func (lr *LeaderboardRoutes) Register(r *gin.Engine) {
	api := r.Group("/api")
	api.GET("/leaderboard", lr.GetLeaderboard)
}

type LeaderboardResponse struct {
	Entries []services.LeaderboardEntry `json:"entries"`
	Total   int64                       `json:"total"`
}

func (lr *LeaderboardRoutes) GetLeaderboard(c *gin.Context) {
	page := c.DefaultQuery("page", "0")
	limit := c.DefaultQuery("limit", "10")

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 0 {
		pageNum = 0
	}

	limitNum, err := strconv.Atoi(limit)
	if err != nil || limitNum < 1 {
		limitNum = 10
	}

	// Get total count
	total, err := lr.leaderboardService.GetTotalEntries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch leaderboard"})
		return
	}

	// Get entries
	entries, err := lr.leaderboardService.GetLeaderboard(limitNum, pageNum*limitNum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch leaderboard"})
		return
	}

	c.JSON(http.StatusOK, LeaderboardResponse{
		Entries: entries,
		Total:   total,
	})
}
