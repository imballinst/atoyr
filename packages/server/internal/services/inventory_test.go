package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserService_AddItems(t *testing.T) {
	db := setupTestDB(t)
	is := NewInventoryService(db)
	ss := NewSessionService(db)
	u := NewUserService(ss)

	user, err := u.CreateUser("test-username")
	assert.NoError(t, err)

	itemIDs := []string{"item1", "item2"}
	userID := user.ID

	err = is.AddItems(itemIDs, userID)
	assert.NoError(t, err)
}

func TestUserService_GetItems(t *testing.T) {
	db := setupTestDB(t)
	ss := NewSessionService(db)
	u := NewUserService(ss)

	is := NewInventoryService(db)

	user, err := u.CreateUser("test-username")
	assert.NoError(t, err)

	itemIDs := []string{"item1", "item2", "item3", "item1"}
	userID := user.ID

	err = is.AddItems(itemIDs, userID)
	assert.NoError(t, err)

	inventoryItems, err := is.GetItems(userID)
	assert.NoError(t, err)
	assert.Len(t, inventoryItems, 3)

	itemQuantities := map[string]int32{}
	for _, item := range inventoryItems {
		itemQuantities[item.ItemID] = item.Quantity
	}

	assert.Equal(t, int32(2), itemQuantities["item1"])
	assert.Equal(t, int32(1), itemQuantities["item2"])
	assert.Equal(t, int32(1), itemQuantities["item3"])
}

func TestUserService_UseItems(t *testing.T) {
	db := setupTestDB(t)
	ss := NewSessionService(db)
	u := NewUserService(ss)

	is := NewInventoryService(db)

	user, err := u.CreateUser("test-username")
	assert.NoError(t, err)

	itemIDs := []string{"item1", "item2", "item3", "item1"}
	userID := user.ID

	err = is.AddItems(itemIDs, userID)
	assert.NoError(t, err)

	inventoryItems, err := is.GetItems(userID)
	assert.NoError(t, err)
	assert.Len(t, inventoryItems, 3)

	itemQuantities := map[string]any{}
	for _, item := range inventoryItems {
		itemQuantities[item.ItemID] = item.Quantity
	}

	assert.Equal(t, int32(2), itemQuantities["item1"])
	assert.Equal(t, int32(1), itemQuantities["item2"])
	assert.Equal(t, int32(1), itemQuantities["item3"])

	err = is.UseItems([]string{"item1", "item2"}, userID)
	assert.NoError(t, err)

	inventoryItems, err = is.GetItems(userID)
	assert.NoError(t, err)
	assert.Len(t, inventoryItems, 2)

	itemQuantities = map[string]any{}
	for _, item := range inventoryItems {
		itemQuantities[item.ItemID] = item.Quantity
	}

	assert.Equal(t, int32(1), itemQuantities["item1"])
	assert.Equal(t, nil, itemQuantities["item2"])
	assert.Equal(t, int32(1), itemQuantities["item3"])
}
