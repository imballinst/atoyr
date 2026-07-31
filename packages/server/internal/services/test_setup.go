package services

import (
	"atoyr/server/internal/testutils"
	"testing"

	"gorm.io/gorm"
)

type TestValues struct {
	DB          *gorm.DB
	Word        *WordService
	Session     *SessionService
	Leaderboard *LeaderboardService
	Game        *GameService
}

func initTestServices(t *testing.T) TestValues {
	db := testutils.SetupTestDB(t)

	ws := initTestWordService()
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	gs := NewGameService(ss, ws, ls, testutils.TestSessionOptions)

	return TestValues{db, ws, ss, ls, gs}
}

const testTopic = "english-words"

func initTestWordService() *WordService {
	return &WordService{
		Words: map[string][]WordDefinition{
			testTopic: {
				{Word: "hello", Definition: "a greeting"},
				{Word: "world", Definition: "the earth"},
				{Word: "apple", Definition: "a fruit"},
				{Word: "banana", Definition: "a yellow fruit"},
				{Word: "cherry", Definition: "a small red fruit"},
				{Word: "dragon", Definition: "a mythical creature"},
				{Word: "elephant", Definition: "a large mammal"},
				{Word: "forest", Definition: "a wooded area"},
				{Word: "guitar", Definition: "a stringed instrument"},
				{Word: "horizon", Definition: "where earth meets sky"},
			},
		},
	}
}
