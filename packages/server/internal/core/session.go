package core

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
	record map[string]int32
}

var SessionDurationManager = sessionDurationManager{
	record: map[string]int32{},
}

func (t sessionDurationManager) Add(sessionID string, duration int32) {
	t.record[sessionID] = duration
}

func (t sessionDurationManager) Extend(sessionID string, value int32) int32 {
	val, ok := t.record[sessionID]
	if !ok {
		return 0
	}

	newRemaining := val + value
	t.record[sessionID] = newRemaining

	return newRemaining
}

func (t sessionDurationManager) Decrement(sessionID string) int32 {
	val, ok := t.record[sessionID]
	if !ok {
		return 0
	}

	newRemaining := val - 1
	t.record[sessionID] = newRemaining

	return newRemaining
}

func (t sessionDurationManager) Get(sessionID string) int32 {
	val, ok := t.record[sessionID]
	if !ok {
		return 0
	}

	return val
}

func (t sessionDurationManager) Clean(sessionID string) {
	delete(t.record, sessionID)
}
