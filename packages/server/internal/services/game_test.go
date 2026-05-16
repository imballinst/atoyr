package services

import (
	"atoyr/server/internal/core"
	"atoyr/server/internal/services/domainmodels"
	"atoyr/server/internal/testutils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGameService_StartGame(t *testing.T) {
	_, itemInfoMap := testutils.SetupItemInfoMap([]string{"item1", "item2"})

	db := testutils.SetupTestDB(t)
	ws := setupTestWordService()
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	gs := NewGameService(ss, ws, ls, itemInfoMap)

	session, _ := ss.Create(false, []string{})
	started, err := gs.StartGame(session.ID)

	assert.NoError(t, err)
	assert.Equal(t, SessionPhasePlaying, started.Phase)
	assert.Equal(t, int32(30), session.RemainingSeconds)
	assert.NotEqual(t, "", started.CurrentWord)
	assert.NotEqual(t, "", started.CurrentWordToken)
}

func TestGameService_StartGame_WithItems(t *testing.T) {
	itemIDs, itemInfoMap := testutils.SetupItemInfoMapWithKind([]string{core.BonusTimerRewardItemID}, []core.ItemInfo{{Kind: core.ItemTimerKind, Value: 10}})

	db := testutils.SetupTestDB(t)
	ws := setupTestWordService()
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

	session.UserEntity = &domainmodels.UserDomain{
		ID: user.ID,
	}
	session.UsedItemIDs = itemIDs

	err = ss.Update(session)
	assert.NoError(t, err)

	session, err = gs.StartGame(session.ID)

	assert.NoError(t, err)
	assert.Equal(t, SessionPhasePlaying, session.Phase)
	assert.Equal(t, int32(40), session.RemainingSeconds)
	assert.NotEqual(t, "", session.CurrentWord)
	assert.NotEqual(t, "", session.CurrentWordToken)
	assert.Equal(t, itemIDs, session.UsedItemIDs)
}

func TestGameService_ContinueGame(t *testing.T) {
	_, itemInfoMap := testutils.SetupItemInfoMap([]string{"item1", "item2"})

	db := testutils.SetupTestDB(t)
	ws := setupTestWordService()
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	gs := NewGameService(ss, ws, ls, itemInfoMap)

	session, _ := ss.Create(false, []string{})
	started, err := gs.StartGame(session.ID)

	assert.NoError(t, err)
	assert.Equal(t, SessionPhasePlaying, started.Phase)
	assert.Equal(t, int32(30), session.RemainingSeconds)
	assert.NotEqual(t, "", started.CurrentWord)
	assert.NotEqual(t, "", started.CurrentWordToken)

	started, err = gs.ContinueGame(session.ID)

	assert.NoError(t, err)
	assert.Equal(t, SessionPhasePlaying, started.Phase)
	assert.NotEqual(t, "", started.CurrentWord)
	assert.NotEqual(t, "", started.CurrentWordToken)
}

func TestGameService_SubmitCorrectAnswer(t *testing.T) {
	_, itemInfoMap := testutils.SetupItemInfoMap([]string{"item1", "item2"})

	db := testutils.SetupTestDB(t)
	ws := setupTestWordService()
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
	assert.Equal(t, int32(1), result.Score)
	assert.Len(t, result.CorrectAttemptTimestamps, 1)
	assert.Len(t, result.CorrectAttemptTimestamps[0], 1)
	assert.Equal(t, int32(1), result.Attempts)
}

func TestGameService_SubmitIncorrectAnswer(t *testing.T) {
	_, itemInfoMap := testutils.SetupItemInfoMap([]string{"item1", "item2"})

	db := testutils.SetupTestDB(t)
	ws := setupTestWordService()
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	gs := NewGameService(ss, ws, ls, itemInfoMap)

	session, _ := ss.Create(false, []string{})
	gs.StartGame(session.ID)

	session, _ = ss.FindByID(session.ID)
	result, err := gs.SubmitAnswer(session.ID, "wronganswer", session.CurrentWordToken)

	assert.NoError(t, err)
	assert.Equal(t, false, result.Correct)
	assert.Equal(t, int32(0), result.Score)
	assert.Len(t, result.CorrectAttemptTimestamps, 0)
	assert.Equal(t, int32(1), result.Attempts)
}

func TestGameService_SubmitCorrectAnswer_AfterCorrectAnswer(t *testing.T) {
	_, itemInfoMap := testutils.SetupItemInfoMap([]string{"item1", "item2"})

	db := testutils.SetupTestDB(t)
	ws := setupTestWordService()
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
	assert.Equal(t, int32(1), result.Score)
	assert.Len(t, result.CorrectAttemptTimestamps, 1)
	assert.Len(t, result.CorrectAttemptTimestamps[0], 1)
	assert.Equal(t, int32(1), result.Attempts)

	// Check session again to get the latest word.
	session, _ = ss.FindByID(session.ID)
	word = session.CurrentWord

	result, err = gs.SubmitAnswer(session.ID, word, session.CurrentWordToken)

	assert.NoError(t, err)
	assert.Equal(t, true, result.Correct)
	assert.Equal(t, int32(2), result.Score)
	assert.Len(t, result.CorrectAttemptTimestamps, 1)
	assert.Len(t, result.CorrectAttemptTimestamps[0], 2)
	assert.Equal(t, int32(2), result.Attempts)
}

func TestGameService_SubmitIncorrectAnswer_AfterCorrectAnswer(t *testing.T) {
	_, itemInfoMap := testutils.SetupItemInfoMap([]string{"item1", "item2"})

	db := testutils.SetupTestDB(t)
	ws := setupTestWordService()
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
	assert.Equal(t, int32(1), result.Score)
	assert.Len(t, result.CorrectAttemptTimestamps, 1)
	assert.Len(t, result.CorrectAttemptTimestamps[0], 1)
	assert.Equal(t, int32(1), result.Attempts)

	// Check session again to get the latest word.
	session, _ = ss.FindByID(session.ID)

	result, err = gs.SubmitAnswer(session.ID, "wronganswer", session.CurrentWordToken)

	assert.NoError(t, err)
	assert.Equal(t, false, result.Correct)
	assert.Equal(t, int32(1), result.Score)
	assert.Len(t, result.CorrectAttemptTimestamps, 2)
	assert.Len(t, result.CorrectAttemptTimestamps[0], 1)
	assert.Len(t, result.CorrectAttemptTimestamps[1], 0)
	assert.Equal(t, int32(2), result.Attempts)
}

func TestGameService_FinishGame(t *testing.T) {
	_, itemInfoMap := testutils.SetupItemInfoMap([]string{"item1", "item2"})

	db := testutils.SetupTestDB(t)
	ws := setupTestWordService()
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	gs := NewGameService(ss, ws, ls, itemInfoMap)

	session, _ := ss.Create(false, []string{})
	_, err := gs.StartGame(session.ID)
	assert.NoError(t, err)

	session.Score = 7
	session.TotalAttempts = 8
	session.CorrectAttemptTimestamps = [][]string{{"ts1"}, {"ts2", "ts3"}}

	err = ss.Update(session)
	assert.NoError(t, err)

	err = gs.FinishGame(session.ID)

	// Check session again to get the latest word.
	session, err = ss.FindByID(session.ID)

	assert.NoError(t, err)
	assert.Len(t, session.CorrectAttemptTimestamps, 2)
	assert.Len(t, session.CorrectAttemptTimestamps[0], 1)
	assert.Len(t, session.CorrectAttemptTimestamps[1], 2)
	assert.Equal(t, int32(7), session.Score)
	assert.Equal(t, int32(8), session.TotalAttempts)
	assert.Equal(t, float32(87.5), session.Accuracy)
}

func TestGameService_TokenGeneration(t *testing.T) {
	_, itemInfoMap := testutils.SetupItemInfoMap([]string{"item1", "item2"})

	db := testutils.SetupTestDB(t)
	ws := setupTestWordService()
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

	db := testutils.SetupTestDB(t)
	ws := setupTestWordService()
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
