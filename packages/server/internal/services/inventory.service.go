package services

import (
	"atoyr/server/internal/models"
	"fmt"
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

func (s *InventoryService) AddItems(itemIDs []string, userID string) error {
	var inventory *models.InventoryEntity

	err := s.db.First(&inventory, "user_entity_id = ?", userID).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if err == gorm.ErrRecordNotFound {
		inventory = &models.InventoryEntity{
			ID:           uuid.New().String(),
			UserEntityID: userID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err = s.db.Save(inventory).Error
		if err != nil {
			return err
		}
	}

	for _, itemID := range itemIDs {
		var inventoryItem *models.InventoryItemEntity

		err = s.db.First(&inventoryItem, "item_id = ? AND inventory_entity_id = ?", itemID, inventory.ID).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		if err == gorm.ErrRecordNotFound {
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

		err = s.db.Save(inventoryItem).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *InventoryService) GetItems(userID string) ([]models.InventoryItemEntity, error) {
	var inventory *models.InventoryEntity

	err := s.db.Find(&models.InventoryEntity{}, "user_entity_id = ?", userID).First(&inventory).Error
	if err != nil {
		return nil, err
	}

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("inventory not found for user %s", userID)
	}

	var inventoryItems []models.InventoryItemEntity

	err = s.db.Find(&models.InventoryItemEntity{}, "inventory_entity_id = ?", inventory.ID).Find(&inventoryItems).Error
	if err != nil {
		return nil, err
	}

	return inventoryItems, nil
}

func (s *InventoryService) UseItems(itemIDs []string, userID string) error {
	if userID == "" {
		// Empty user ID means no user, so we can skip inventory operations.
		return nil
	}

	var inventory *models.InventoryEntity

	err := s.db.First(&inventory, "user_entity_id = ?", userID).Error
	if err != nil {
		return err
	}

	for _, itemID := range itemIDs {
		var inventoryItem *models.InventoryItemEntity

		err = s.db.First(&inventoryItem, "item_id = ? AND inventory_entity_id = ?", itemID, inventory.ID).Error
		if err != nil {
			return err
		}

		inventoryItem.UpdatedAt = time.Now()
		inventoryItem.Quantity -= 1

		if inventoryItem.Quantity > 0 {
			err = s.db.Save(inventoryItem).Error
		} else {
			err = s.db.Delete(inventoryItem).Error
		}
	}

	return err
}
