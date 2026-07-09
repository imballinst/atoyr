package api

import (
	"atoyr/server/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterHealthRoutes(rg *gin.Engine, healthService *services.HealthService) {
	rg.GET("/api/v1/health", func(c *gin.Context) {
		health := healthService.GetHealth()

		status := http.StatusOK
		if !health.Dependencies.Database {
			status = http.StatusInternalServerError
		}

		c.JSON(status, health)
	})
}
