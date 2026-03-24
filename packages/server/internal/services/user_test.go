package services

import (
	"atoyr/server/internal/core"
	"atoyr/server/internal/testutils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserService_UpsertUserFromSession(t *testing.T) {
	db := testutils.SetupTestDB(t)
	ss := NewSessionService(db)
	is := NewInventoryService(db, core.ItemInfoMap{})
	s := NewUserService(ss, is)

	session, _ := ss.Create(false, []string{})
	session.Phase = "finished"

	err := ss.Update(session)
	assert.NoError(t, err)

	user, err := s.UpsertUserFromSession("testuser", session.ID)
	assert.NoError(t, err)
	assert.NotNil(t, user)
}

func TestUserService_FindUserByID(t *testing.T) {
	db := testutils.SetupTestDB(t)
	ss := NewSessionService(db)
	is := NewInventoryService(db, core.ItemInfoMap{})
	s := NewUserService(ss, is)

	user, err := s.CreateUser("testuser")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	persisted, err := s.FindUserByID(user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, persisted)
	assert.Equal(t, user.ID, persisted.ID)
}
