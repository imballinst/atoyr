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
	ID        string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	EndsAt    time.Time
	Topic     string
	Mode      string
	// Available phases: idle, playing, finished.
	Phase                    string
	Score                    int32
	TotalAttempts            int32
	Accuracy                 float32
	DurationSeconds          int32
	CorrectAttemptTimestamps JSON `gorm:"type:jsonb"`
	AutoVoice                bool
	UsedWords                pq.StringArray `gorm:"type:text"`
	CurrentWord              string
	CurrentScrambledWord     string
	CurrentWordDefinition    string
	CurrentWordToken         string
	UsedItemIDs              pq.StringArray `gorm:"type:text"`
}
