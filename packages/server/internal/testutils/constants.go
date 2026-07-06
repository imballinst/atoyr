package testutils

import (
	"atoyr/server/internal/core"
	"time"
)

var TestSessionOptions = core.SessionOptions{
	Duration: 1,
	Tick:     0.5,
}

func NowMockFn() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
}
