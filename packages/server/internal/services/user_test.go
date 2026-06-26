package services

import (
	"atoyr/server/internal/core"
	"atoyr/server/internal/testutils"
	"atoyr/server/internal/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserService_UpsertUserFromSession(t *testing.T) {
	db := testutils.SetupTestDB(t)
	ss := NewSessionService(db)
	is := NewInventoryService(db, core.ItemInfoMap{})
	s := NewUserService(ss, is)

	session, _ := ss.Create(false, []string{}, testutils.TestSessionOptions.Duration)
	session.Phase = core.SessionPhaseFinished

	err := ss.Update(session)
	assert.NoError(t, err)

	user, err := s.UpsertUserFromSession(utils.GenerateUUID(), session.ID)
	assert.NoError(t, err)
	assert.NotNil(t, user)
}

func TestUserService_FindUserByID(t *testing.T) {
	db := testutils.SetupTestDB(t)
	ss := NewSessionService(db)
	is := NewInventoryService(db, core.ItemInfoMap{})
	s := NewUserService(ss, is)

	user, err := s.CreateUser(utils.GenerateUUID())
	assert.NoError(t, err)
	assert.NotNil(t, user)

	persisted, err := s.FindUserByID(user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, persisted)
	assert.Equal(t, user.ID, persisted.ID)
}
