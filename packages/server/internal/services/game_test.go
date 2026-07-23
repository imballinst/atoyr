package services

import (
	"atoyr/server/internal/core"
	"atoyr/server/internal/testutils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGameService_StartGame(t *testing.T) {
	testServices := initTestServices(t)

	session, _ := testServices.Session.Create(false, []string{}, "vanilla", testutils.TestSessionOptions.Duration+5)
	started, err := testServices.Game.StartGame(session.ID, "test-device-id", false)

	assert.NoError(t, err)
	assert.Equal(t, core.SessionPhasePlaying, started.Phase)
	assert.Equal(t, "vanilla", started.Mode)
	assert.Equal(t, int32(6), session.DurationSeconds)
	assert.NotEqual(t, "", started.CurrentWord)
	assert.NotEqual(t, "", started.CurrentWordToken)
	assert.NotEqual(t, "", started.CurrentWordDefinition)
}

func TestGameService_ContinueGame(t *testing.T) {
	testServices := initTestServices(t)

	session, _ := testServices.Session.Create(false, []string{}, "vanilla", testutils.TestSessionOptions.Duration+5)
	started, err := testServices.Game.StartGame(session.ID, "test-device-id", false)

	assert.NoError(t, err)
	assert.Equal(t, core.SessionPhasePlaying, started.Phase)
	assert.Equal(t, int32(6), session.DurationSeconds)
	assert.NotEqual(t, "", started.CurrentWord)
	assert.NotEqual(t, "", started.CurrentWordToken)

	started, err = testServices.Game.ContinueGame(session.ID)

	assert.NoError(t, err)
	assert.Equal(t, core.SessionPhasePlaying, started.Phase)
	assert.NotEqual(t, "", started.CurrentWord)
	assert.NotEqual(t, "", started.CurrentWordToken)
}

func TestGameService_SubmitCorrectAnswer(t *testing.T) {
	for _, mode := range []string{"vanilla", "blind"} {
		t.Run(mode, func(t *testing.T) {
			testServices := initTestServices(t)

			session, _ := testServices.Session.Create(false, []string{}, mode, testutils.TestSessionOptions.Duration+5)
			testServices.Game.StartGame(session.ID, "test-device-id", false)

			state := testServices.Store.Get(session.ID)
			word := state.CurrentWord

			result, err := testServices.Game.SubmitAnswer(session.ID, word, state.CurrentWordToken)

			assert.NoError(t, err)
			assert.Equal(t, true, result.Correct)
			assert.Equal(t, int32(1), result.Score)
			assert.Len(t, result.CorrectAttemptTimestamps, 1)
			assert.Len(t, result.CorrectAttemptTimestamps[0], 1)
			assert.Equal(t, int32(1), result.Attempts)
			if mode == "blind" {
				assert.Equal(t, "", result.ScrambledWordDefinition)
			} else {
				assert.NotEqual(t, "", result.ScrambledWordDefinition)
			}
		})
	}
}

func TestGameService_SubmitIncorrectAnswer(t *testing.T) {
	for _, mode := range []string{"vanilla", "blind"} {
		t.Run(mode, func(t *testing.T) {
			testServices := initTestServices(t)

			session, _ := testServices.Session.Create(false, []string{}, mode, testutils.TestSessionOptions.Duration+5)
			testServices.Game.StartGame(session.ID, "test-device-id", false)

			state := testServices.Store.Get(session.ID)
			originalEndsAt := state.EndsAt
			result, err := testServices.Game.SubmitAnswer(session.ID, "wronganswer", state.CurrentWordToken)

			assert.NoError(t, err)
			assert.Equal(t, false, result.Correct)
			assert.Equal(t, int32(0), result.Score)
			assert.Len(t, result.CorrectAttemptTimestamps, 0)
			assert.Equal(t, int32(1), result.Attempts)

			state = testServices.Store.Get(session.ID)
			assert.True(t, state.EndsAt.Before(originalEndsAt))
		})
	}
}

func TestGameService_SubmitCorrectAnswer_AfterCorrectAnswer(t *testing.T) {
	testServices := initTestServices(t)

	session, _ := testServices.Session.Create(false, []string{}, "vanilla", testutils.TestSessionOptions.Duration+5)
	testServices.Game.StartGame(session.ID, "test-device-id", false)

	state := testServices.Store.Get(session.ID)
	word := state.CurrentWord

	result, err := testServices.Game.SubmitAnswer(session.ID, word, state.CurrentWordToken)

	assert.NoError(t, err)
	assert.Equal(t, true, result.Correct)
	assert.Equal(t, int32(1), result.Score)
	assert.Len(t, result.CorrectAttemptTimestamps, 1)
	assert.Len(t, result.CorrectAttemptTimestamps[0], 1)
	assert.Equal(t, int32(1), result.Attempts)

	// Check session again to get the latest word.
	state = testServices.Store.Get(session.ID)
	word = state.CurrentWord

	result, err = testServices.Game.SubmitAnswer(session.ID, word, state.CurrentWordToken)

	assert.NoError(t, err)
	assert.Equal(t, true, result.Correct)
	assert.Equal(t, int32(2), result.Score)
	assert.Len(t, result.CorrectAttemptTimestamps, 1)
	assert.Len(t, result.CorrectAttemptTimestamps[0], 2)
	assert.Equal(t, int32(2), result.Attempts)
}

func TestGameService_SubmitIncorrectAnswer_AfterCorrectAnswer(t *testing.T) {
	testServices := initTestServices(t)

	session, _ := testServices.Session.Create(false, []string{}, "vanilla", testutils.TestSessionOptions.Duration+5)
	testServices.Game.StartGame(session.ID, "test-device-id", false)

	state := testServices.Store.Get(session.ID)
	word := state.CurrentWord

	result, err := testServices.Game.SubmitAnswer(session.ID, word, state.CurrentWordToken)

	assert.NoError(t, err)
	assert.Equal(t, true, result.Correct)
	assert.Equal(t, int32(1), result.Score)
	assert.Len(t, result.CorrectAttemptTimestamps, 1)
	assert.Len(t, result.CorrectAttemptTimestamps[0], 1)
	assert.Equal(t, int32(1), result.Attempts)

	// Check session again to get the latest word.
	state = testServices.Store.Get(session.ID)

	result, err = testServices.Game.SubmitAnswer(session.ID, "wronganswer", state.CurrentWordToken)

	assert.NoError(t, err)
	assert.Equal(t, false, result.Correct)
	assert.Equal(t, int32(1), result.Score)
	assert.Len(t, result.CorrectAttemptTimestamps, 2)
	assert.Len(t, result.CorrectAttemptTimestamps[0], 1)
	assert.Len(t, result.CorrectAttemptTimestamps[1], 0)
	assert.Equal(t, int32(2), result.Attempts)
}

func TestGameService_FinishGame(t *testing.T) {
	testServices := initTestServices(t)

	session, _ := testServices.Session.Create(false, []string{}, "vanilla", testutils.TestSessionOptions.Duration+5)
	_, err := testServices.Game.StartGame(session.ID, "test-device-id", false)
	assert.NoError(t, err)

	state := testServices.Store.Get(session.ID)
	state.Score = 7
	state.TotalAttempts = 8
	state.CorrectAttemptTimestamps = [][]string{{"ts1"}, {"ts2", "ts3"}}
	testServices.Store.Set(state)

	err = testServices.Game.FinishGame(session.ID)

	assert.NoError(t, err)
	assert.Nil(t, testServices.Store.Get(session.ID))

	// Verify the SQLite leaderboard fallback has the finished summary.
	finished, err := testServices.Leaderboard.GetLeaderboard("vanilla", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, finished, 1)
	assert.Equal(t, int32(7), finished[0].Score)
	assert.Equal(t, int32(8), finished[0].TotalAttempts)
	assert.Equal(t, float32(87.5), finished[0].Accuracy)
}

func TestGameService_StartGame_BlindMode(t *testing.T) {
	testServices := initTestServices(t)

	session, _ := testServices.Session.Create(false, []string{}, "blind", testutils.TestSessionOptions.Duration+5)
	started, err := testServices.Game.StartGame(session.ID, "test-device-id", false)

	assert.NoError(t, err)
	assert.Equal(t, core.SessionPhasePlaying, started.Phase)
	assert.Equal(t, "blind", started.Mode)
	assert.NotEqual(t, "", started.CurrentWord)
	assert.Equal(t, "", started.CurrentWordDefinition)
}

func TestGameService_ContinueGame_BlindMode(t *testing.T) {
	testServices := initTestServices(t)

	session, _ := testServices.Session.Create(false, []string{}, "blind", testutils.TestSessionOptions.Duration+5)
	_, err := testServices.Game.StartGame(session.ID, "test-device-id", false)
	assert.NoError(t, err)

	continued, err := testServices.Game.ContinueGame(session.ID)

	assert.NoError(t, err)
	assert.Equal(t, core.SessionPhasePlaying, continued.Phase)
	assert.Equal(t, "", continued.CurrentWordDefinition)
}

func TestGameService_TokenGeneration(t *testing.T) {
	testServices := initTestServices(t)

	// Same word should generate same token
	token1 := testServices.Game.generateToken("test")
	token2 := testServices.Game.generateToken("test")

	assert.Equal(t, token1, token2)

	// Different words should generate different tokens
	token3 := testServices.Game.generateToken("test2")

	assert.NotEqual(t, token1, token3)
}

func TestGameService_AttemptCounting(t *testing.T) {
	testServices := initTestServices(t)

	session, _ := testServices.Session.Create(false, []string{}, "vanilla", testutils.TestSessionOptions.Duration+5)
	testServices.Game.StartGame(session.ID, "test-device-id", false)

	testServices.Game.SubmitAnswer(session.ID, "wronganswer", testServices.Store.Get(session.ID).CurrentWordToken)
	state := testServices.Store.Get(session.ID)
	assert.Equal(t, int32(1), state.TotalAttempts)

	testServices.Game.SubmitAnswer(session.ID, "wronganswer2", testServices.Store.Get(session.ID).CurrentWordToken)
	state = testServices.Store.Get(session.ID)
	assert.Equal(t, int32(2), state.TotalAttempts)
}
