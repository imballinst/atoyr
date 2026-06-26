package main

import (
	"fmt"
	"log"
	"os"

	api "atoyr/server/internal/api"
	"atoyr/server/internal/core"
	"atoyr/server/internal/database"
	"atoyr/server/internal/middleware"
	"atoyr/server/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting server...")

	err := godotenv.Load(".env.local")
	if err != nil {
		log.Println("Error loading .env.local file")
	}

	// Set environment
	if os.Getenv("ENV") == "" {
		os.Setenv("ENV", "development")
	}

	log.Println("Initializing database...")

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

	sessionOptions := core.SessionOptions{
		Duration: 30,
		Tick:     1,
	}

	sessionService := services.NewSessionService(db)
	leaderboardService := services.NewLeaderboardService(db)
	gameService := services.NewGameService(sessionService, wordService, leaderboardService, sessionOptions)

	// Initialize Gin
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORSMiddleware())

	server := api.NewServer(gameService, sessionService, leaderboardService, sessionOptions)
	api.RegisterHandlers(router, server)

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
