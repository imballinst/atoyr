package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// MetricSnapshot stores time-series metrics data
type MetricSnapshot struct {
	ID              uint `gorm:"primaryKey"`
	Timestamp       time.Time
	RequestCount    int64   `gorm:"column:request_count"`
	ErrorCount4xx   int64   `gorm:"column:error_count_4xx"`
	ErrorCount5xx   int64   `gorm:"column:error_count_5xx"`
	ActiveGames     int64   `gorm:"column:active_games"`
	ResponseTimeP50 int64   `gorm:"column:response_time_p50"` // milliseconds
	ResponseTimeP95 int64   `gorm:"column:response_time_p95"` // milliseconds
	ResponseTimeP99 int64   `gorm:"column:response_time_p99"` // milliseconds
	MemoryUsageMb   float64 `gorm:"column:memory_usage_mb"`
}

func (MetricSnapshot) TableName() string {
	return "metrics_snapshots"
}

type JSON json.RawMessage

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	case nil:
		*j = JSON(nil)
		return nil
	default:
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

// SessionEntity is the minimal SQLite registry used for crash recovery and
// admin counting. Gameplay state lives in the in-memory SessionStore and is
// flushed to AGS Cloud Save asynchronously.
type SessionEntity struct {
	ID        string `gorm:"primaryKey"`
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
	EndsAt    time.Time
	Mode      string
	// Available phases: idle, playing, finished.
	Phase string
	// CurrentWord is kept in the registry so the SSE finish event can report
	// the last word without requiring the rich in-memory state to survive.
	CurrentWord string
}

// LeaderboardSessionEntity maps to the same session_entities table but includes
// the gameplay fields required by the SQLite leaderboard fallback. This model is
// intentionally separate from the minimal registry so that the game service
// cannot accidentally read or write gameplay fields during normal gameplay.
type LeaderboardSessionEntity struct {
	ID                       string `gorm:"primaryKey"`
	UserID                   string
	CreatedAt                time.Time
	UpdatedAt                time.Time
	EndsAt                   time.Time
	Mode                     string
	Phase                    string
	Score                    int32
	TotalAttempts            int32
	Accuracy                 float32
	DurationSeconds          int32
	CorrectAttemptTimestamps JSON `gorm:"type:jsonb"`
	AutoVoice                bool
	UsedWords                pq.StringArray `gorm:"type:text"`
	WordDefinitions          pq.StringArray `gorm:"type:text"`
	CurrentWord              string
	CurrentScrambledWord     string
	CurrentWordDefinition    string
	CurrentWordToken         string
	UsedItemIDs              pq.StringArray `gorm:"type:text"`
}

func (LeaderboardSessionEntity) TableName() string {
	return "session_entities"
}
