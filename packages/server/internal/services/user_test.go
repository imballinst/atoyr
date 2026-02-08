package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserService_UpsertUserFromSession(t *testing.T) {
	db := setupTestDB(t)
	ss := NewSessionService(db)
	s := NewUserService(ss)

	session, _ := ss.Create(false)
	session.Phase = "finished"

	err := ss.Update(session)
	assert.NoError(t, err)

	user, err := s.UpsertUserFromSession("testuser", session.ID)
	assert.NoError(t, err)
	assert.NotNil(t, user)
}
