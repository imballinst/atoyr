package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserService_UpsertUserFromSession(t *testing.T) {
	db := setupTestDB(t)
	ss := NewSessionService(db)
	s := NewUserService(ss)

	session, _ := ss.Create(false, []string{})
	session.Phase = "finished"

	err := ss.Update(session)
	assert.NoError(t, err)

	user, err := s.UpsertUserFromSession("testuser", session.ID)
	assert.NoError(t, err)
	assert.NotNil(t, user)
}

func TestUserService_FindUserByID(t *testing.T) {
	db := setupTestDB(t)
	ss := NewSessionService(db)
	s := NewUserService(ss)

	user, err := s.CreateUser("testuser")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	persisted, err := s.FindUserByID(user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, persisted)
	assert.Equal(t, user.ID, persisted.ID)
}
