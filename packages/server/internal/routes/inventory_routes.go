package routes

import (
	"log"
	"net/http"

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
	Data []InventoryItem `json:"data"`
}

type InventoryItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Quantity int32  `json:"quantity"`
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

	items, err := ir.inventoryService.GetInventoryItems(userId)
	if err != nil {
		log.Println("Failed to get inventory items:", err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get inventory items, " + err.Error()})
		return
	}

	data := []InventoryItem{}

	for _, item := range items {
		data = append(data, InventoryItem{
			ID: item.ItemID,
			// TODO: create items service to get name.
			Name:     "stub name",
			Quantity: int32(item.Quantity),
		})
	}

	c.JSON(http.StatusOK, GetItemsResponse{
		Data: data,
	})
}
