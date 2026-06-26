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

func initTestWordService() *WordService {
	return &WordService{
		Words: []WordDefinition{
			{Word: "hello"},
			{Word: "world"},
			{Word: "apple"},
			{Word: "banana"},
			{Word: "cherry"},
			{Word: "dragon"},
			{Word: "elephant"},
			{Word: "forest"},
			{Word: "guitar"},
			{Word: "horizon"},
		},
	}
}
