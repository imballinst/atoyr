package api

import (
	"atoyr/server/internal/core"
	"atoyr/server/internal/services"
)

type Server struct {
	gameService        *services.GameService
	sessionService     *services.SessionService
	inventoryService   *services.InventoryService
	leaderboardService *services.LeaderboardService
	userService        *services.UserService
	sessionOptions     core.SessionOptions
}

func NewServer(
	gameService *services.GameService,
	sessionService *services.SessionService,
	inventoryService *services.InventoryService,
	leaderboardService *services.LeaderboardService,
	userService *services.UserService,
	sessionOptions core.SessionOptions,
) *Server {
	return &Server{
		gameService:        gameService,
		sessionService:     sessionService,
		inventoryService:   inventoryService,
		leaderboardService: leaderboardService,
		userService:        userService,
		sessionOptions:     sessionOptions,
	}
}
