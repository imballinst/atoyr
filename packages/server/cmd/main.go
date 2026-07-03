package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
	metricsCollector := middleware.NewMetricsCollector()

	// Initialize auth middleware
	allowedEmails := strings.Split(os.Getenv("ALLOWED_ADMIN_EMAILS"), ",")
	authMiddleware, err := middleware.NewAuthMiddleware(allowedEmails)
	if err != nil {
		log.Fatalf("Failed to initialize auth middleware: %v", err)
	}

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
	statsService := services.NewStatsService(db, metricsCollector)

	restoreActiveSessions(gameService, sessionService)

	// Start metrics snapshot worker (writes to DB every 60s for time series)
	metricsCollector.StartSnapshotWorker(db, 60*time.Second)

	// Clean up snapshots older than 3 months on startup
	metricsCollector.CleanupOldSnapshots(db, 3*30*24*time.Hour)

	// Initialize Gin
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORSMiddleware())
	api.RegisterAdminRoutes(router, statsService, authMiddleware)

	router.Use(middleware.Metrics(metricsCollector))

	server := api.NewServer(gameService, sessionService, leaderboardService, sessionOptions)
	api.RegisterHandlers(router, server)

	router.Static("/dashboard", "./web/static")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	address := fmt.Sprintf(":%s", port)
	srv := &http.Server{
		Addr:    address,
		Handler: router,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server on %s", address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func restoreActiveSessions(gameService *services.GameService, sessionService *services.SessionService) {
	fmt.Println("[restore active sessions] finding sessions...")

	sessions, err := sessionService.FindPlaying()
	if err != nil {
		log.Printf("Failed to find play2ing sessions: %v", err)
		return
	}

	fmt.Println("[restore active sessions] restoring sessions...")

	restored := 0
	for _, s := range sessions {
		fmt.Printf("[restore active sessions] restoring session %s...\n", s.ID)

		remaining := int32(time.Until(s.EndsAt).Seconds())
		if remaining <= 0 {
			go gameService.FinishGame(s.ID)
			continue
		}
		go gameService.RestoreTimer(s.ID, remaining)
		restored++
	}

	log.Printf("Restored %d active game timers\n", restored)
}
