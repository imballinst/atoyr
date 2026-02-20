package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionService_Create(t *testing.T) {
	db := setupTestDB(t)
	service := NewSessionService(db)

	session, err := service.Create(true, []string{"test-item-id"})
	assert.NoError(t, err)
	assert.NotEqual(t, "", session.ID)
	assert.Equal(t, "idle", session.Phase)
	assert.Equal(t, int32(0), session.Score)
	assert.Equal(t, int32(30), session.RemainingSeconds)
	assert.Equal(t, true, session.AutoVoice)
	assert.Equal(t, "test-item-id", session.UsedItemIDs[0])
}

func TestSessionService_FindByID(t *testing.T) {
	db := setupTestDB(t)
	service := NewSessionService(db)

	// Create session
	created, err := service.Create(false, []string{})
	assert.NoError(t, err)

	// Find session
	found, err := service.FindByID(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestSessionService_SetCurrentWord(t *testing.T) {
	db := setupTestDB(t)
	service := NewSessionService(db)

	session, _ := service.Create(false, []string{})

	err := service.SetCurrentWord(session.ID, "test", "definition", "token123", []string{"test"})
	assert.NoError(t, err)

	found, _ := service.FindByID(session.ID)
	assert.Equal(t, "test", found.CurrentWord)
	assert.Equal(t, "token123", found.CurrentWordToken)
	assert.Len(t, found.UsedWords, 1)
	assert.Equal(t, "test", found.UsedWords[0])
}

func TestSessionService_EndSession(t *testing.T) {
	db := setupTestDB(t)
	service := NewSessionService(db)

	session, _ := service.Create(false, []string{})

	service.EndSession(session.ID)
	found, _ := service.FindByID(session.ID)

	assert.Equal(t, "finished", found.Phase)
	assert.Equal(t, int32(0), found.RemainingSeconds)
}
