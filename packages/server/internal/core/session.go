package core

type SessionOptions struct {
	Duration int32
	Tick     float32
}

const BonusDurationPerWordWithAutoVoice = 5

type sessionDurationManager struct {
	record map[string]int32
}

var SessionDurationManager = sessionDurationManager{
	record: map[string]int32{},
}

func (t sessionDurationManager) Add(sessionID string, duration int32) {
	t.record[sessionID] = duration
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
