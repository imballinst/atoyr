package main

import (
	"fmt"
	"log"
	"os"

	"atoyr/server/internal/database"
	"atoyr/server/internal/middleware"
	"atoyr/server/internal/routes"
	"atoyr/server/internal/services"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set environment
	if os.Getenv("NODE_ENV") == "" {
		os.Setenv("NODE_ENV", "development")
	}

	// Initialize database
	db, err := database.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize services
	wordService, err := services.NewWordService()
	if err != nil {
		log.Fatalf("Failed to initialize word service: %v", err)
	}

	sessionService := services.NewSessionService(db)
	leaderboardService := services.NewLeaderboardService(db)
	gameService := services.NewGameService(sessionService, wordService, leaderboardService)

	// Initialize Gin
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORSMiddleware())

	// Register routes
	gameRoutes := routes.NewGameRoutes(gameService, sessionService)
	gameRoutes.Register(router)

	leaderboardRoutes := routes.NewLeaderboardRoutes(leaderboardService)
	leaderboardRoutes.Register(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	address := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on %s", address)
	if err := router.Run(address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
