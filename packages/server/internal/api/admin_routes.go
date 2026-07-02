package api

import (
	"atoyr/server/internal/middleware"
	"atoyr/server/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(rg *gin.Engine, statsService *services.StatsService, auth *middleware.AuthMiddleware) {
	admin := rg.Group("/api/v1/admin")
	admin.Use(auth.RequireAuth())
	{
		admin.GET("/stats", func(c *gin.Context) {
			stats, err := statsService.GetStats()
			if err != nil {
				c.JSON(500, gin.H{"error": "failed to get stats"})
				return
			}
			c.JSON(200, stats)
		})

		admin.GET("/timeseries", func(c *gin.Context) {
			period := c.DefaultQuery("period", "24h")
			granularity := c.DefaultQuery("granularity", "")

			timeseries, err := statsService.GetTimeSeries(period, granularity)
			if err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, timeseries)
		})
	}
}
