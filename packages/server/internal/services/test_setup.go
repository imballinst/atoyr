package services

import (
	"atoyr/server/internal/core"
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
	Inventory   *InventoryService
	User        *UserService
}

func initTestServices(t *testing.T, itemInfoMap core.ItemInfoMap) TestValues {
	db := testutils.SetupTestDB(t)

	ws := initTestWordService()
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	is := NewInventoryService(db, itemInfoMap)
	gs := NewGameService(ss, ws, ls, is, itemInfoMap, testutils.TestSessionOptions)
	us := NewUserService(ss, is)

	return TestValues{db, ws, ss, ls, gs, is, us}
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
