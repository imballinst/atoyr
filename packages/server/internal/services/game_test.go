package services

import (
	"atoyr/server/internal/core"
	"atoyr/server/internal/testutils"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestGameService_StartGame(t *testing.T) {
	_, itemInfoMap := testutils.SetupItemInfoMap([]string{"item1", "item2"})

	db := setupTestDB(t)
	ws := setupTestWordService(t)
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	gs := NewGameService(ss, ws, ls, itemInfoMap)

	session, _ := ss.Create(false, []string{})
	started, err := gs.StartGame(session.ID)

	assert.NoError(t, err)
	assert.Equal(t, sessionPhasePlaying, started.Phase)
	assert.Equal(t, int32(30), session.RemainingSeconds)
	assert.NotEqual(t, "", started.CurrentWord)
	assert.NotEqual(t, "", started.CurrentWordToken)
}

func TestGameService_StartGame_WithItems(t *testing.T) {
	itemIDs, itemInfoMap := testutils.SetupItemInfoMapWithKind([]string{core.BonusTimerRewardItemID}, []core.ItemInfo{{Kind: core.ItemTimerKind, Value: 10}})

	db := setupTestDB(t)
	ws := setupTestWordService(t)
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	gs := NewGameService(ss, ws, ls, itemInfoMap)
	is := NewInventoryService(db, itemInfoMap)
	us := NewUserService(ss, is)

	user, err := us.CreateUser("test")
	assert.NoError(t, err)

	err = is.AddItems(itemIDs, user.ID)
	assert.NoError(t, err)

	session, err := ss.Create(false, []string{})
	assert.NoError(t, err)

	session.UserEntityID = &user.ID
	session.UsedItemIDs = itemIDs

	err = ss.Update(session)
	assert.NoError(t, err)

	session, err = gs.StartGame(session.ID)

	assert.NoError(t, err)
	assert.Equal(t, sessionPhasePlaying, session.Phase)
	assert.Equal(t, int32(40), session.RemainingSeconds)
	assert.NotEqual(t, "", session.CurrentWord)
	assert.NotEqual(t, "", session.CurrentWordToken)
	assert.Equal(t, pq.StringArray(itemIDs), session.UsedItemIDs)
}

func TestGameService_SubmitCorrectAnswer(t *testing.T) {
	_, itemInfoMap := testutils.SetupItemInfoMap([]string{"item1", "item2"})

	db := setupTestDB(t)
	ws := setupTestWordService(t)
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	gs := NewGameService(ss, ws, ls, itemInfoMap)

	session, _ := ss.Create(false, []string{})
	gs.StartGame(session.ID)

	session, _ = ss.FindByID(session.ID)
	word := session.CurrentWord

	result, err := gs.SubmitAnswer(session.ID, word, session.CurrentWordToken)

	assert.NoError(t, err)
	assert.Equal(t, true, result.Correct)
	assert.NotEqual(t, int32(0), result.Score)
	assert.Len(t, result.CorrectAttemptTimestamps, 1)
	assert.Len(t, result.CorrectAttemptTimestamps[0], 1)
	assert.Equal(t, int32(1), result.Attempts)
}

func TestGameService_SubmitIncorrectAnswer(t *testing.T) {
	_, itemInfoMap := testutils.SetupItemInfoMap([]string{"item1", "item2"})

	db := setupTestDB(t)
	ws := setupTestWordService(t)
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	gs := NewGameService(ss, ws, ls, itemInfoMap)

	session, _ := ss.Create(false, []string{})
	gs.StartGame(session.ID)

	result, err := gs.SubmitAnswer(session.ID, "wronganswer", session.CurrentWordToken)

	assert.NoError(t, err)
	assert.Equal(t, true, result.Correct)
	assert.Equal(t, int32(0), result.Score)
	assert.Len(t, result.CorrectAttemptTimestamps, 0)
	assert.Equal(t, int32(1), result.Attempts)
}

func TestGameService_TokenGeneration(t *testing.T) {
	_, itemInfoMap := testutils.SetupItemInfoMap([]string{"item1", "item2"})

	db := setupTestDB(t)
	ws := setupTestWordService(t)
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	gs := NewGameService(ss, ws, ls, itemInfoMap)

	// Same word should generate same token
	token1 := gs.generateToken("test")
	token2 := gs.generateToken("test")

	assert.Equal(t, token1, token2)

	// Different words should generate different tokens
	token3 := gs.generateToken("test2")

	assert.NotEqual(t, token1, token3)
}

func TestGameService_AttemptCounting(t *testing.T) {
	_, itemInfoMap := testutils.SetupItemInfoMap([]string{"item1", "item2"})

	db := setupTestDB(t)
	ws := setupTestWordService(t)
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	gs := NewGameService(ss, ws, ls, itemInfoMap)

	session, _ := ss.Create(false, []string{})
	gs.StartGame(session.ID)

	gs.SubmitAnswer(session.ID, "wronganswer", session.CurrentWordToken)
	session, _ = ss.FindByID(session.ID)

	assert.Equal(t, int32(1), session.TotalAttempts)

	gs.SubmitAnswer(session.ID, "wronganswer2", session.CurrentWordToken)
	session, _ = ss.FindByID(session.ID)

	assert.Equal(t, int32(2), session.TotalAttempts)
}
