package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type JSON json.RawMessage

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = JSON(result)
	return err
}

// Value return json value, implement driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// SessionEntity represents an active game session
type SessionEntity struct {
	ID        string    `gorm:"primaryKey;type:text"`
	CreatedAt time.Time `gorm:"type:datetime"`
	UpdatedAt time.Time `gorm:"type:datetime"`
	EndsAt    time.Time `gorm:"type:datetime"`
	// Available phases: idle, playing, finished.
	Phase                    string         `gorm:"type:text;default:idle"`
	Score                    int32          `gorm:"type:integer;default:0"`
	TotalAttempts            int32          `gorm:"type:integer;default:0"`
	Accuracy                 float32        `gorm:"type:real;default:0"`
	RemainingSeconds         int32          `gorm:"type:integer"`
	CorrectAttemptTimestamps JSON           `gorm:"type:jsonb;default:'[]'"`
	AutoVoice                bool           `gorm:"type:boolean;default:false"`
	UsedWords                pq.StringArray `gorm:"type:text"`
	WordDefinitions          pq.StringArray `gorm:"type:text"`
	CurrentWord              string         `gorm:"type:text"`
	CurrentScrambledWord     string         `gorm:"type:text"`
	CurrentWordDefinition    string         `gorm:"type:text"`
	CurrentWordToken         string         `gorm:"type:text"`
	UsedItemIDs              pq.StringArray `gorm:"type:text"`
	// Belongs to user entity.
	UserEntityID *string `gorm:"type:text;index:idx_sessions_user"`
	UserEntity   *UserEntity
}

type UserEntity struct {
	ID        string    `gorm:"primaryKey;type:text"`
	CreatedAt time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	Username  string    `gorm:"type:text;uniqueIndex:idx_profiles_username"`
}

type InventoryEntity struct {
	ID        string    `gorm:"primaryKey;type:text"`
	CreatedAt time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	// Belongs to user entity.
	UserEntityID string `gorm:"type:text;index:idx_inventory_user"`
	UserEntity   UserEntity
}

type InventoryItemEntity struct {
	ItemID    string    `gorm:"primaryKey;type:text"`
	Quantity  int32     `gorm:"type:integer;default:0"`
	CreatedAt time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	// Belongs to inventory and user entity.
	InventoryEntityID string `gorm:"type:text;index:idx_inventory_items_inventory"`
	InventoryEntity   InventoryEntity
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
	s.UsedWords = state.UsedWords
	s.CurrentWordToken = state.CurrentWordToken
}
