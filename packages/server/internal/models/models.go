package models

import (
	"time"

	"github.com/lib/pq"
)

// SessionEntity represents an active game session
type SessionEntity struct {
	ID                       string         `gorm:"primaryKey;type:text"`
	CreatedAt                time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	ExpiresAt                time.Time      `gorm:"type:datetime;index:idx_sessions_expires"`
	Phase                    string         `gorm:"type:text;default:idle"`
	Score                    int32          `gorm:"type:integer;default:0"`
	TotalAttempts            int32          `gorm:"type:integer;default:0"`
	RemainingSeconds         int32          `gorm:"type:integer"`
	CorrectAttemptTimestamps pq.StringArray `gorm:"type:text;default:'[]'"`
	AutoVoice                bool           `gorm:"type:boolean;default:false"`
	UsedWords                pq.StringArray `gorm:"type:text;default:'[]'"`
	WordDefinitions          pq.StringArray `gorm:"type:text;default:'[]'"`
	CurrentWord              string         `gorm:"type:text"`
	CurrentWordDefinition    string         `gorm:"type:text"`
	CurrentWordToken         string         `gorm:"type:text"`
}

// ResultEntity represents a completed game result
type ResultEntity struct {
	ID            string         `gorm:"primaryKey;type:text"`
	SessionID     string         `gorm:"type:text;index:idx_results_session"`
	Timestamp     time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP;index:idx_results_timestamp"`
	Score         int32          `gorm:"type:integer;index:idx_results_score"`
	TotalAttempts int32          `gorm:"type:integer"`
	RewardIDs     pq.StringArray `gorm:"type:text;default:'[]'"`
	Accuracy      float32        `gorm:"type:real"`
	DurationMs    int32          `gorm:"type:integer"`
}

// GameSessionState represents the current state of a game
type GameSessionState struct {
	Phase            string   `json:"phase"`
	Score            int32    `json:"score"`
	TotalAttempts    int32    `json:"totalAttempts"`
	RemainingSeconds int32    `json:"remainingSeconds"`
	UsedWords        []string `json:"usedWords"`
	CurrentWordToken string   `json:"currentWordToken"`
}

func (s *SessionEntity) GetState() GameSessionState {
	usedWords := []string(s.UsedWords)

	return GameSessionState{
		Phase:            s.Phase,
		Score:            s.Score,
		TotalAttempts:    s.TotalAttempts,
		RemainingSeconds: s.RemainingSeconds,
		UsedWords:        usedWords,
		CurrentWordToken: s.CurrentWordToken,
	}
}

func (s *SessionEntity) SetState(state GameSessionState) {
	s.Phase = state.Phase
	s.Score = state.Score
	s.TotalAttempts = state.TotalAttempts
	s.RemainingSeconds = state.RemainingSeconds
	s.UsedWords = pq.StringArray(state.UsedWords)
	s.CurrentWordToken = state.CurrentWordToken
}
