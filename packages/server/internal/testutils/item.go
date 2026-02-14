package testutils

import (
	"atoyr/server/internal/core"
	"fmt"
)

func SetupItemInfoMap(itemIDs []string) ([]string, core.ItemInfoMap) {
	itemInfoMap := core.ItemInfoMap{}

	for _, itemID := range itemIDs {
		itemInfoMap[itemID] = core.ItemInfo{
			ID:          itemID,
			Kind:        fmt.Sprintf("Kind for %s", itemID),
			Name:        fmt.Sprintf("Name for %s", itemID),
			Description: fmt.Sprintf("Description for %s", itemID),
			Value:       0,
		}
	}

	return itemIDs, itemInfoMap
}

func GetMockInventoryItem(inventoryItemMap map[string]core.InventoryItem, itemInfoMap core.ItemInfoMap, itemID string) core.InventoryItem {
	itemInfo := itemInfoMap[itemID]
	inventoryItem := inventoryItemMap[itemID]

	return core.InventoryItem{
		ItemID:      itemID,
		Name:        itemInfo.Name,
		Description: itemInfo.Description,
		Quantity:    inventoryItem.Quantity,
	}
}
