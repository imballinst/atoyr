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
	started, err := testServices.Game.StartGame(session.ID, "test-device-id")

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
	started, err := testServices.Game.StartGame(session.ID, "test-device-id")

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
			testServices.Game.StartGame(session.ID, "test-device-id")

			session, _ = testServices.Session.FindByID(session.ID)
			word := session.CurrentWord

			result, err := testServices.Game.SubmitAnswer(session.ID, word, session.CurrentWordToken)

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
			testServices.Game.StartGame(session.ID, "test-device-id")

			session, _ = testServices.Session.FindByID(session.ID)
			originalEndsAt := session.EndsAt
			result, err := testServices.Game.SubmitAnswer(session.ID, "wronganswer", session.CurrentWordToken)

			assert.NoError(t, err)
			assert.Equal(t, false, result.Correct)
			assert.Equal(t, int32(0), result.Score)
			assert.Len(t, result.CorrectAttemptTimestamps, 0)
			assert.Equal(t, int32(1), result.Attempts)

			session, _ = testServices.Session.FindByID(session.ID)
			assert.True(t, session.EndsAt.Before(originalEndsAt))
		})
	}
}

func TestGameService_SubmitCorrectAnswer_AfterCorrectAnswer(t *testing.T) {
	testServices := initTestServices(t)

	session, _ := testServices.Session.Create(false, []string{}, "vanilla", testutils.TestSessionOptions.Duration+5)
	testServices.Game.StartGame(session.ID, "test-device-id")

	session, _ = testServices.Session.FindByID(session.ID)
	word := session.CurrentWord

	result, err := testServices.Game.SubmitAnswer(session.ID, word, session.CurrentWordToken)

	assert.NoError(t, err)
	assert.Equal(t, true, result.Correct)
	assert.Equal(t, int32(1), result.Score)
	assert.Len(t, result.CorrectAttemptTimestamps, 1)
	assert.Len(t, result.CorrectAttemptTimestamps[0], 1)
	assert.Equal(t, int32(1), result.Attempts)

	// Check session again to get the latest word.
	session, _ = testServices.Session.FindByID(session.ID)
	word = session.CurrentWord

	result, err = testServices.Game.SubmitAnswer(session.ID, word, session.CurrentWordToken)

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
	testServices.Game.StartGame(session.ID, "test-device-id")

	session, _ = testServices.Session.FindByID(session.ID)
	word := session.CurrentWord

	result, err := testServices.Game.SubmitAnswer(session.ID, word, session.CurrentWordToken)

	assert.NoError(t, err)
	assert.Equal(t, true, result.Correct)
	assert.Equal(t, int32(1), result.Score)
	assert.Len(t, result.CorrectAttemptTimestamps, 1)
	assert.Len(t, result.CorrectAttemptTimestamps[0], 1)
	assert.Equal(t, int32(1), result.Attempts)

	// Check session again to get the latest word.
	session, _ = testServices.Session.FindByID(session.ID)

	result, err = testServices.Game.SubmitAnswer(session.ID, "wronganswer", session.CurrentWordToken)

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
	_, err := testServices.Game.StartGame(session.ID, "test-device-id")
	assert.NoError(t, err)

	session.Score = 7
	session.TotalAttempts = 8
	session.CorrectAttemptTimestamps = [][]string{{"ts1"}, {"ts2", "ts3"}}

	err = testServices.Session.Update(session)
	assert.NoError(t, err)

	err = testServices.Game.FinishGame(session.ID)

	// Check session again to get the latest word.
	session, err = testServices.Session.FindByID(session.ID)

	assert.NoError(t, err)
	assert.Len(t, session.CorrectAttemptTimestamps, 2)
	assert.Len(t, session.CorrectAttemptTimestamps[0], 1)
	assert.Len(t, session.CorrectAttemptTimestamps[1], 2)
	assert.Equal(t, int32(7), session.Score)
	assert.Equal(t, int32(8), session.TotalAttempts)
	assert.Equal(t, float32(87.5), session.Accuracy)
}

func TestGameService_StartGame_BlindMode(t *testing.T) {
	testServices := initTestServices(t)

	session, _ := testServices.Session.Create(false, []string{}, "blind", testutils.TestSessionOptions.Duration+5)
	started, err := testServices.Game.StartGame(session.ID, "test-device-id")

	assert.NoError(t, err)
	assert.Equal(t, core.SessionPhasePlaying, started.Phase)
	assert.Equal(t, "blind", started.Mode)
	assert.NotEqual(t, "", started.CurrentWord)
	assert.Equal(t, "", started.CurrentWordDefinition)
}

func TestGameService_ContinueGame_BlindMode(t *testing.T) {
	testServices := initTestServices(t)

	session, _ := testServices.Session.Create(false, []string{}, "blind", testutils.TestSessionOptions.Duration+5)
	_, err := testServices.Game.StartGame(session.ID, "test-device-id")
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
	testServices.Game.StartGame(session.ID, "test-device-id")

	testServices.Game.SubmitAnswer(session.ID, "wronganswer", session.CurrentWordToken)
	session, _ = testServices.Session.FindByID(session.ID)

	assert.Equal(t, int32(1), session.TotalAttempts)

	testServices.Game.SubmitAnswer(session.ID, "wronganswer2", session.CurrentWordToken)
	session, _ = testServices.Session.FindByID(session.ID)

	assert.Equal(t, int32(2), session.TotalAttempts)
}
