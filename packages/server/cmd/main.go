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

	sessionService := services.NewSessionService(db)
	leaderboardService := services.NewLeaderboardService(db)
	gameService, err := services.NewGameService(sessionService, wordService, leaderboardService)
	inventoryService := services.NewInventoryService(db)

	if err != nil {
		log.Fatalf("Failed to initialize game service: %v", err)
	}

	// Initialize Gin
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORSMiddleware())

	// Register routes
	gameRoutes := routes.NewGameRoutes(gameService, sessionService, inventoryService)
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
