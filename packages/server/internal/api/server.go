package api

import (
	"atoyr/server/internal/core"
	"atoyr/server/internal/services"
)

type Server struct {
	gameService        *services.GameService
	sessionService     *services.SessionService
	leaderboardService *services.LeaderboardService
	sessionOptions     core.SessionOptions
}

func NewServer(
	gameService *services.GameService,
	sessionService *services.SessionService,
	leaderboardService *services.LeaderboardService,
	sessionOptions core.SessionOptions,
) *Server {
	return &Server{
		gameService:        gameService,
		sessionService:     sessionService,
		leaderboardService: leaderboardService,
		sessionOptions:     sessionOptions,
	}
}
