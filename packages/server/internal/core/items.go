package core

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	BonusTimerRewardItemID = "938d5ff9fda94b8098b02cc891093d10"
	ItemTimerKind          = "timer"
)

type ItemInfoMap map[string]ItemInfo

type ItemInfo struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Value       int32  `json:"value"`
}

func LoadItems() (ItemInfoMap, error) {
	itemsPath := os.Getenv("ITEMS_PATH")
	if itemsPath == "" {
		return nil, fmt.Errorf("ITEMS_PATH environment variable is not set")
	}

	data, err := os.ReadFile(itemsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read items file: %w", err)
	}

	var dbContent []ItemInfo

	if err := json.Unmarshal(data, &dbContent); err != nil {
		return nil, fmt.Errorf("failed to parse items file: %w", err)
	}

	itemMap := ItemInfoMap{}
	for _, item := range dbContent {
		itemMap[item.ID] = item
	}

	return itemMap, nil
}

type InventoryItem struct {
	ItemID      string
	Name        string
	Description string
	Quantity    int32
}
