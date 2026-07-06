package core

import "sync"

const (
	SessionPhasePlaying  = "playing"
	SessionPhaseFinished = "finished"
)

type SessionOptions struct {
	Duration int32
	Tick     float32
}

const BonusDurationPerWordWithAutoVoice = 5

func GetSessionPhase(remainingSeconds int32) string {
	if remainingSeconds <= 0 {
		return SessionPhaseFinished
	}

	return SessionPhasePlaying
}

// Session duration manager.
type sessionDurationManager struct {
	mu     sync.RWMutex
	record map[string]int32
}

var SessionDurationManager = &sessionDurationManager{
	record: map[string]int32{},
}

func (t *sessionDurationManager) Add(sessionID string, duration int32) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.record[sessionID] = duration
}

func (t *sessionDurationManager) Extend(sessionID string, value int32) int32 {
	t.mu.Lock()
	defer t.mu.Unlock()
	val, ok := t.record[sessionID]
	if !ok {
		return 0
	}

	newRemaining := val + value
	t.record[sessionID] = newRemaining

	return newRemaining
}

func (t *sessionDurationManager) Decrement(sessionID string) int32 {
	t.mu.Lock()
	defer t.mu.Unlock()
	val, ok := t.record[sessionID]
	if !ok {
		return 0
	}

	newRemaining := val - 1
	t.record[sessionID] = newRemaining

	return newRemaining
}

func (t *sessionDurationManager) Get(sessionID string) int32 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	val, ok := t.record[sessionID]
	if !ok {
		return 0
	}

	return val
}

func (t *sessionDurationManager) Clean(sessionID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.record, sessionID)
}
