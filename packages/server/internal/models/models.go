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
	ID              uint      `gorm:"primaryKey"`
	Timestamp       time.Time `gorm:"index"`
	RequestCount    int64     `gorm:"column:request_count;not null;default:0"`
	ErrorCount4xx   int64     `gorm:"column:error_count_4xx;not null;default:0"`
	ErrorCount5xx   int64     `gorm:"column:error_count_5xx;not null;default:0"`
	ActiveGames     int64     `gorm:"column:active_games;not null;default:0"`
	ResponseTimeP50 int64     `gorm:"column:response_time_p50;not null;default:0"` // milliseconds
	ResponseTimeP95 int64     `gorm:"column:response_time_p95;not null;default:0"` // milliseconds
	ResponseTimeP99 int64     `gorm:"column:response_time_p99;not null;default:0"` // milliseconds
	MemoryUsageMb   float64   `gorm:"column:memory_usage_mb;not null;default:0"`
}

func (MetricSnapshot) TableName() string {
	return "metrics_snapshots"
}

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
	Phase                    string         `gorm:"index:idx_phase_score,priority:1;type:text;default:idle"`
	Score                    int32          `gorm:"index:idx_phase_score,priority:2;type:integer;default:0"`
	TotalAttempts            int32          `gorm:"type:integer;default:0"`
	Accuracy                 float32        `gorm:"type:real;default:0"`
	DurationSeconds          int32          `gorm:"type:integer"`
	CorrectAttemptTimestamps JSON           `gorm:"type:jsonb;default:'[]'"`
	AutoVoice                bool           `gorm:"type:boolean;default:false"`
	UsedWords                pq.StringArray `gorm:"type:text"`
	WordDefinitions          pq.StringArray `gorm:"type:text"`
	CurrentWord              string         `gorm:"type:text"`
	CurrentScrambledWord     string         `gorm:"type:text"`
	CurrentWordDefinition    string         `gorm:"type:text"`
	CurrentWordToken         string         `gorm:"type:text"`
	UsedItemIDs              pq.StringArray `gorm:"type:text"`
}
