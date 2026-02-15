package routes

import (
	"log"
	"net/http"

	"atoyr/server/internal/core"
	"atoyr/server/internal/services"

	"github.com/gin-gonic/gin"
)

type InventoryRoutes struct {
	inventoryService *services.InventoryService
}

func NewInventoryRoutes(inventoryService *services.InventoryService, sessionService *services.SessionService) *InventoryRoutes {
	return &InventoryRoutes{
		inventoryService: inventoryService,
	}
}

func (ir *InventoryRoutes) Register(r *gin.Engine) {
	api := r.Group("/api")
	inventory := api.Group("/inventory")

	inventory.GET("/v1/items/me", ir.GetItems)
}

type GetItemsResponse struct {
	Data []core.InventoryItem `json:"data"`
}

func (ir *InventoryRoutes) GetItems(c *gin.Context) {
	userId, err := c.Cookie("user_id")
	if err == http.ErrNoCookie {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err != nil {
		log.Println("Failed to get user_id cookie:", err)

		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get user_id cookie, " + err.Error()})
		return
	}

	items, err := ir.inventoryService.GetItems(userId)
	if err != nil {
		log.Println("Failed to get inventory items:", err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get inventory items, " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetItemsResponse{
		Data: items,
	})
}
