package services

import (
	"atoyr/server/internal/core"
	"atoyr/server/internal/testutils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserService_AddItems(t *testing.T) {
	itemIDs, itemInfoMap := testutils.SetupItemInfoMap([]string{"item1", "item2"})

	db := testutils.SetupTestDB(t)
	is := NewInventoryService(db, itemInfoMap)
	ss := NewSessionService(db)
	u := NewUserService(ss, is)

	user, err := u.CreateUser("test-username")
	assert.NoError(t, err)

	userID := user.ID

	err = is.AddItems(itemIDs, userID)
	assert.NoError(t, err)
}

func TestUserService_GetItems(t *testing.T) {
	itemIDs, itemInfoMap := testutils.SetupItemInfoMap([]string{"item1", "item2", "item3", "item1"})

	db := testutils.SetupTestDB(t)
	ss := NewSessionService(db)
	is := NewInventoryService(db, itemInfoMap)
	u := NewUserService(ss, is)

	user, err := u.CreateUser("test-username")
	assert.NoError(t, err)

	userID := user.ID

	err = is.AddItems(itemIDs, userID)
	assert.NoError(t, err)

	inventoryItems, err := is.GetItems(userID)
	assert.NoError(t, err)
	assert.Len(t, inventoryItems, 3)

	itemMap := map[string]core.InventoryItem{}
	for _, item := range inventoryItems {
		itemMap[item.ItemID] = item
	}

	assert.Equal(t, testutils.GetMockInventoryItem(itemMap, itemInfoMap, "item1"), itemMap["item1"])
	assert.Equal(t, testutils.GetMockInventoryItem(itemMap, itemInfoMap, "item2"), itemMap["item2"])
	assert.Equal(t, testutils.GetMockInventoryItem(itemMap, itemInfoMap, "item3"), itemMap["item3"])
}

func TestUserService_UseItems(t *testing.T) {
	itemIDs, itemInfoMap := testutils.SetupItemInfoMap([]string{"item1", "item2", "item3", "item1"})

	db := testutils.SetupTestDB(t)
	ss := NewSessionService(db)
	is := NewInventoryService(db, itemInfoMap)
	u := NewUserService(ss, is)

	user, err := u.CreateUser("test-username")
	assert.NoError(t, err)

	userID := user.ID

	err = is.AddItems(itemIDs, userID)
	assert.NoError(t, err)

	inventoryItems, err := is.GetItems(userID)
	assert.NoError(t, err)
	assert.Len(t, inventoryItems, 3)

	itemMap := map[string]core.InventoryItem{}
	for _, item := range inventoryItems {
		itemMap[item.ItemID] = item
	}

	assert.Equal(t, testutils.GetMockInventoryItem(itemMap, itemInfoMap, "item1"), itemMap["item1"])
	assert.Equal(t, testutils.GetMockInventoryItem(itemMap, itemInfoMap, "item2"), itemMap["item2"])
	assert.Equal(t, testutils.GetMockInventoryItem(itemMap, itemInfoMap, "item3"), itemMap["item3"])

	err = is.UseItems([]string{"item1", "item2"}, userID)
	assert.NoError(t, err)

	inventoryItems, err = is.GetItems(userID)
	assert.NoError(t, err)
	assert.Len(t, inventoryItems, 2)

	itemMap = map[string]core.InventoryItem{}
	for _, item := range inventoryItems {
		itemMap[item.ItemID] = item
	}

	emptyInventoryItemStock := testutils.GetMockInventoryItem(nil, nil, "")

	assert.Equal(t, testutils.GetMockInventoryItem(itemMap, itemInfoMap, "item1"), itemMap["item1"])
	assert.Equal(t, emptyInventoryItemStock, itemMap["item2"])
	assert.Equal(t, testutils.GetMockInventoryItem(itemMap, itemInfoMap, "item3"), itemMap["item3"])
}
