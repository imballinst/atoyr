package core

import (
	"atoyr/server/internal/services/domainmodels"
	"sync"
)

// SessionStore holds the rich gameplay state in memory. It is keyed by session ID
// and is safe for concurrent access.
type SessionStore struct {
	sessions sync.Map
}

// NewSessionStore creates a new in-memory session store.
func NewSessionStore() *SessionStore {
	return &SessionStore{}
}

// Get returns the session with the given id, or nil if it is not present.
func (s *SessionStore) Get(id string) *domainmodels.SessionDomain {
	val, ok := s.sessions.Load(id)
	if !ok {
		return nil
	}
	return val.(*domainmodels.SessionDomain)
}

// Set stores the given session. If session is nil, this is a no-op.
func (s *SessionStore) Set(session *domainmodels.SessionDomain) {
	if session == nil {
		return
	}
	s.sessions.Store(session.ID, session)
}

// Delete removes the session with the given id from the store.
func (s *SessionStore) Delete(id string) {
	s.sessions.Delete(id)
}

// GlobalSessionStore is the package-level store used by the production server.
// It is intentionally global so that crash-recovery timers can operate without
// needing the store reference from the service layer.
var GlobalSessionStore = NewSessionStore()
