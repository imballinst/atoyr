package services

import (
	"atoyr/server/internal/models"
	"time"

	"github.com/google/uuid"
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

func (s *InventoryService) AddItemsToInventory(itemIDs []string, userID string) error {
	var inventory *models.InventoryEntity

	err := s.db.Find(&models.InventoryEntity{}, "user_entity_id = ?", userID).First(&inventory).Error
	if err != nil {
		return err
	}

	if inventory == nil {
		inventory = &models.InventoryEntity{
			ID:           uuid.New().String(),
			UserEntityID: userID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err = s.db.Model(&models.InventoryEntity{}).Save(inventory).Error
		if err != nil {
			return err
		}
	}

	for _, itemID := range itemIDs {
		var inventoryItem *models.InventoryItemEntity

		err = s.db.Find(models.InventoryItemEntity{}, "item_id = ?", itemID).First(&inventoryItem).Error
		if err != nil {
			return err
		}

		if inventoryItem == nil {
			inventoryItem = &models.InventoryItemEntity{
				InventoryEntityID: inventory.ID,
				ItemID:            itemID,
				Quantity:          0,
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			}
		}

		inventoryItem.UpdatedAt = time.Now()
		inventoryItem.Quantity += 1

		err = s.db.Model(&models.InventoryItemEntity{}).Save(inventoryItem).Error
		if err != nil {
			return err
		}
	}

	return nil
}
