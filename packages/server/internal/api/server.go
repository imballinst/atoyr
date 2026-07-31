package api

import (
	"atoyr/server/internal/core"
	"atoyr/server/internal/services"
)

type Server struct {
	gameService        *services.GameService
	sessionService     *services.SessionService
	leaderboardService *services.LeaderboardService
	wordService        *services.WordService
	sessionOptions     core.SessionOptions
}

func NewServer(
	gameService *services.GameService,
	sessionService *services.SessionService,
	leaderboardService *services.LeaderboardService,
	wordService *services.WordService,
	sessionOptions core.SessionOptions,
) *Server {
	return &Server{
		gameService:        gameService,
		sessionService:     sessionService,
		leaderboardService: leaderboardService,
		wordService:        wordService,
		sessionOptions:     sessionOptions,
	}
}
