package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (gr *Server) GetApiV1ItemsMe(c *gin.Context) {
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

	items, err := gr.inventoryService.GetItems(userId)
	if err != nil {
		log.Println("Failed to get inventory items:", err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get inventory items, " + err.Error()})
		return
	}

	apiItems := make([]InventoryItem, len(items))
	for i, item := range items {
		apiItems[i] = toApiInventoryItem(item)
	}

	c.JSON(http.StatusOK, GetItemsResponse{
		Data: apiItems,
	})
}
