package testutils

import (
	"atoyr/server/internal/core"
	"fmt"
	"reflect"
)

func SetupItemInfoMap(itemIDs []string) ([]string, core.ItemInfoMap) {
	return SetupItemInfoMapWithKind(itemIDs, []core.ItemInfo{})
}

func SetupItemInfoMapWithKind(itemIDs []string, partialInfos []core.ItemInfo) ([]string, core.ItemInfoMap) {
	itemInfoMap := core.ItemInfoMap{}

	for i, itemID := range itemIDs {
		kind := fmt.Sprintf("Kind for %s", itemID)
		name := fmt.Sprintf("Name for %s", itemID)
		description := fmt.Sprintf("Description for %s", itemID)
		value := int32(0)

		if i+1 <= len(partialInfos) {
			partialInfo := partialInfos[i]

			kind = getNonZeroValue(kind, partialInfo.Kind)
			name = getNonZeroValue(name, partialInfo.Name)
			value = getNonZeroValue(value, partialInfo.Value)
		}

		itemInfoMap[itemID] = core.ItemInfo{
			ID:          itemID,
			Kind:        kind,
			Name:        name,
			Description: description,
			Value:       value,
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

func getNonZeroValue[T any](val1, val2 T) T {
	if reflect.ValueOf(val2).IsZero() {
		return val1
	}

	return val2
}
