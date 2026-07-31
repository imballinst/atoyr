package services

import (
	"atoyr/server/internal/core"
	"atoyr/server/internal/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSessionService_Create(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewSessionService(db)

	session, err := service.Create(true, []string{"test-item-id"}, "vanilla", testTopic, testutils.TestSessionOptions.Duration)
	assert.NoError(t, err)
	assert.NotEqual(t, "", session.ID)
	assert.Equal(t, "idle", session.Phase)
	assert.Equal(t, int32(0), session.Score)
	assert.Equal(t, int32(6), session.DurationSeconds)
	assert.Equal(t, true, session.AutoVoice)
	assert.Equal(t, "test-item-id", session.UsedItemIDs[0])
	assert.Equal(t, "english-words", session.Topic)
}

func TestSessionService_FindByID(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewSessionService(db)

	// Create session
	created, err := service.Create(false, []string{}, "vanilla", testTopic, testutils.TestSessionOptions.Duration)
	assert.NoError(t, err)

	// Find session
	found, err := service.FindByID(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestSessionService_EndSession(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewSessionService(db)

	session, _ := service.Create(false, []string{}, "vanilla", testTopic, testutils.TestSessionOptions.Duration)

	service.EndSession(session.ID)
	found, _ := service.FindByID(session.ID)

	assert.Equal(t, core.SessionPhaseFinished, found.Phase)
	// Duration should be kept as-is.
	assert.Equal(t, int32(1), found.DurationSeconds)
}

func TestSessionService_EndSession_DoubleEnd(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewSessionService(db)

	session, _ := service.Create(false, []string{}, "vanilla", testTopic, testutils.TestSessionOptions.Duration)

	// First call should succeed.
	err := service.EndSession(session.ID)
	assert.NoError(t, err)
	found1, _ := service.FindByID(session.ID)
	assert.Equal(t, core.SessionPhaseFinished, found1.Phase)

	// Wait a bit so updated_at would differ if a second write happened.
	time.Sleep(2 * time.Millisecond)

	// Second call should be a no-op (no error, phase stays finished, updated_at unchanged).
	err = service.EndSession(session.ID)
	assert.NoError(t, err)
	found2, _ := service.FindByID(session.ID)
	assert.Equal(t, core.SessionPhaseFinished, found2.Phase)
	assert.Equal(t, found1.UpdatedAt.UnixNano(), found2.UpdatedAt.UnixNano())
}
