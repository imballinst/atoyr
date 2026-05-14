package api

import "atoyr/server/internal/services"

type Server struct {
	gameService      *services.GameService
	sessionService   *services.SessionService
	inventoryService *services.InventoryService
	userService      *services.UserService
}

func NewServer(
	gameService *services.GameService,
	sessionService *services.SessionService,
	inventoryService *services.InventoryService,
	userService *services.UserService,
) *Server {
	return &Server{
		gameService:      gameService,
		sessionService:   sessionService,
		inventoryService: inventoryService,
		userService:      userService,
	}
}
