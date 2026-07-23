package eventhandler

import "time"

type SessionStarted struct {
	EventName string    `json:"eventName"`
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id"`
	Mode      string    `json:"mode"`
	StartedAt time.Time `json:"started_at"`
}

type SessionFinished struct {
	EventName     string    `json:"eventName"`
	SessionID     string    `json:"session_id"`
	UserID        string    `json:"user_id"`
	Mode          string    `json:"mode"`
	Score         int       `json:"score"`
	TotalAttempts int       `json:"total_attempts"`
	Accuracy      float64   `json:"accuracy"`
	FinishedAt    time.Time `json:"finished_at"`
}

type RoundFinished SessionFinished
