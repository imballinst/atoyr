package services

import (
	"gorm.io/gorm"
)

type InventoryService struct {
	db *gorm.DB
}

func NewInventoryService(db *gorm.DB) *InventoryService {
	return &InventoryService{
		db: db,
	}
}

func (u *InventoryService) AddItemsToInventory(itemIDs []string, userID string) error {
	// TODO: steps:
	// 1. Get the user
	// 2. Append the item IDs
	// 3. Save into Inventory model

	return nil
}
