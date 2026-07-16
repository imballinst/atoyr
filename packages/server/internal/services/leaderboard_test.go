package services

import (
	"math/rand"
	"testing"
	"time"

	"atoyr/server/internal/core"
	"atoyr/server/internal/models"
	"atoyr/server/internal/testutils"
	"atoyr/server/internal/utils"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestLeaderboardService_GetLeaderboard(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewLeaderboardService(db)

	// Add some test results
	results := []models.SessionEntity{
		{
			ID:                       "session1",
			Mode:                     "vanilla",
			Score:                    100,
			TotalAttempts:            10,
			Accuracy:                 90,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		},
		{
			ID:                       "session2",
			Mode:                     "vanilla",
			Score:                    150,
			TotalAttempts:            15,
			Accuracy:                 95,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		},
		{
			ID:                       "session3",
			Mode:                     "vanilla",
			Score:                    80,
			TotalAttempts:            8,
			Accuracy:                 85.12345,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		},
		{
			ID:                       "session4",
			Mode:                     "vanilla",
			Score:                    0,
			TotalAttempts:            8,
			Accuracy:                 0,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		},
	}

	for _, result := range results {
		db.Create(&result)
	}

	// Get leaderboard
	entries, err := service.GetLeaderboard("vanilla", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, entries, 3)

	// Verify sorting by score (descending)
	assert.Equal(t, int32(150), entries[0].Score)
	assert.Equal(t, int32(100), entries[1].Score)
	assert.Equal(t, int32(80), entries[2].Score)

	// Truncates to last 2 decimals.
	assert.Equal(t, float32(85.12), entries[2].Accuracy)
}

func TestLeaderboardService_GetLeaderboard_SameScore(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewLeaderboardService(db)

	// Add some test results
	results := []models.SessionEntity{
		{
			ID:                       "s01",
			Mode:                     "vanilla",
			Score:                    100,
			TotalAttempts:            10,
			Accuracy:                 90,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		},
		{
			ID:                       "s02",
			Mode:                     "vanilla",
			Score:                    100,
			TotalAttempts:            9,
			Accuracy:                 95,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		},
	}

	for _, result := range results {
		db.Create(&result)
	}

	// Get leaderboard
	entries, err := service.GetLeaderboard("vanilla", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, entries, 2)

	// Verify sorting by score, then accuracy (descending)
	assert.Equal(t, "s02", entries[0].ID)
	assert.Equal(t, float32(95), entries[0].Accuracy)

	assert.Equal(t, "s01", entries[1].ID)
	assert.Equal(t, float32(90), entries[1].Accuracy)
}

func TestLeaderboardService_GetLeaderboard_DifferentModes(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewLeaderboardService(db)

	// Add some test results
	results := []models.SessionEntity{
		{
			ID:                       "s01",
			Mode:                     "vanilla",
			Score:                    100,
			TotalAttempts:            10,
			Accuracy:                 90,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		},
		{
			ID:                       "s02",
			Mode:                     "blind",
			Score:                    100,
			TotalAttempts:            9,
			Accuracy:                 95,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		},
	}

	for _, result := range results {
		db.Create(&result)
	}

	// Get leaderboard
	entries, err := service.GetLeaderboard("vanilla", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	assert.Equal(t, "s01", entries[0].ID)
	assert.Equal(t, float32(90), entries[0].Accuracy)

	// Get leaderboard, blind mode
	entries, err = service.GetLeaderboard("blind", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	assert.Equal(t, "s02", entries[0].ID)
	assert.Equal(t, float32(95), entries[0].Accuracy)
}

func TestLeaderboardService_Pagination(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewLeaderboardService(db)

	// Add 25 results
	for i := range 25 {
		result := models.SessionEntity{
			ID:                       "session" + string(rune(i)),
			Mode:                     "vanilla",
			Score:                    int32(100 + i),
			TotalAttempts:            int32(10),
			Accuracy:                 90,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		}
		db.Create(&result)
	}

	// Get first page (limit 10)
	entries, _ := service.GetLeaderboard("vanilla", 10, 0)
	assert.Len(t, entries, 10)

	// Get second page
	entries, _ = service.GetLeaderboard("vanilla", 10, 10)
	assert.Len(t, entries, 10)

	// Get third page (should have 5)
	entries, _ = service.GetLeaderboard("vanilla", 10, 20)
	assert.Len(t, entries, 5)
}

func TestLeaderboardService_GetTotalEntries(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewLeaderboardService(db)

	total, _ := service.GetTotalEntries("vanilla")
	assert.Equal(t, int64(0), total)

	// Add 5 results
	for i := range 5 {
		result := models.SessionEntity{
			ID:                       "session" + string(rune(i)),
			Mode:                     "vanilla",
			Score:                    int32(100 + i),
			TotalAttempts:            10,
			Accuracy:                 90,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		}
		db.Create(&result)
	}

	total, _ = service.GetTotalEntries("vanilla")
	assert.Equal(t, int64(5), total)
}

func TestLeaderboardService_GetPercentile(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewLeaderboardService(db)

	results := []models.SessionEntity{}

	for i := range 10 {
		result := models.SessionEntity{
			ID:                       utils.GenerateUUID(),
			Mode:                     "vanilla",
			Score:                    int32(i * 10),
			TotalAttempts:            int32(10),
			Accuracy:                 float32(100 - (i * 10)),
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			WordDefinitions:          pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		}
		db.Create(&result)
		results = append(results, result)
	}

	// Create a single blind session with a high score.
	blindSession := models.SessionEntity{
		ID:                       utils.GenerateUUID(),
		Mode:                     "blind",
		Score:                    100,
		TotalAttempts:            int32(10),
		Accuracy:                 100,
		CorrectAttemptTimestamps: models.JSON{},
		UsedWords:                pq.StringArray{},
		WordDefinitions:          pq.StringArray{},
		UsedItemIDs:              pq.StringArray{},
		Phase:                    core.SessionPhaseFinished,
	}
	db.Create(&blindSession)

	for i := 1; i < 10; i++ {
		// 0 is not eligible in the code, so we start from 1 to 10.
		id := results[i].ID
		score := results[i].Score
		totalEligible := float32(8)
		totalBelowCurrentScore := float32(i) - 1
		expectedPercentile := (totalBelowCurrentScore / totalEligible) * 100

		percentile, err := service.GetPercentile(id, "vanilla", score)
		assert.NoError(t, err)
		assert.Equal(t, expectedPercentile, percentile, map[string]any{"score": score, "totalBelowCurrentScore": totalBelowCurrentScore, "totalEligible": totalEligible})
	}

	// A blind session should only be compared against other blind sessions.
	percentile, err := service.GetPercentile(blindSession.ID, "blind", blindSession.Score)
	assert.NoError(t, err)
	assert.Equal(t, float32(0), percentile)
}

// TestLeaderboardService_ExcludesSessionWithUnfinishedPhase reproduces the bug where a
// session that was not properly finished (phase remains "playing" instead of "finished")
// is excluded from the leaderboard, resulting in fewer entries than expected.
//
// Root cause: runTimer() in game.service.go exits without calling FinishGame() when
// SessionDurationManager already reads 0 at the start of a loop iteration. This happens
// when the last Decrement() comes from a wrong answer (updateSessionBasedOnAnswerResult)
// rather than from the timer's own Decrement call. The session stays in "playing" phase
// in the DB and the leaderboard query WHERE phase = "finished" excludes it.
func TestLeaderboardService_ExcludesSessionWithUnfinishedPhase(t *testing.T) {
	db := testutils.SetupTestDB(t)
	leaderboardService := NewLeaderboardService(db)
	sessionService := NewSessionService(db)

	// Create 3 finished sessions with scores
	for i := range 3 {
		session, err := sessionService.Create(false, []string{}, "vanilla", 30)
		assert.NoError(t, err)

		session.Score = int32(30 - i*10)
		session.TotalAttempts = 10
		session.Accuracy = utils.CalculateAccuracy(session.Score, session.TotalAttempts)
		session.Phase = core.SessionPhaseFinished
		assert.NoError(t, sessionService.Update(session))
	}

	// Player's session — finished in terms of score, but phase was never set to "finished".
	// This simulates the runTimer race condition where FinishGame() is not called.
	playerSession, err := sessionService.Create(false, []string{}, "vanilla", 30)
	assert.NoError(t, err)

	playerSession.Score = 5
	playerSession.TotalAttempts = 10
	playerSession.Accuracy = utils.CalculateAccuracy(5, 10)
	playerSession.Phase = core.SessionPhasePlaying // NOT set to "finished" — the bug
	assert.NoError(t, sessionService.Update(playerSession))

	// Fetch leaderboard
	entries, err := leaderboardService.GetLeaderboard("vanilla", 10, 0)
	assert.NoError(t, err)

	total, err := leaderboardService.GetTotalEntries("vanilla")
	assert.NoError(t, err)

	t.Logf("Leaderboard returned %d entries (total=%d), expected 4", len(entries), total)
	for i, e := range entries {
		t.Logf("  [%d] id=%s score=%d", i+1, e.ID, e.Score)
	}

	// BUG: only 3 entries are returned because the player's session is excluded
	assert.Len(t, entries, 3, "BUG: expected 4 entries but got 3 — the player's session was excluded because FinishGame() was never called")
	assert.Equal(t, int64(3), total, "BUG: expected total 4 but got 3")
}

// TestLeaderboardService_IncludesFinishedSession verifies that when FinishGame IS called
// (phase = "finished"), the player's session is included in the leaderboard.
func TestLeaderboardService_IncludesFinishedSession(t *testing.T) {
	db := testutils.SetupTestDB(t)
	leaderboardService := NewLeaderboardService(db)
	sessionService := NewSessionService(db)

	// Create 3 finished sessions with scores
	for i := range 3 {
		session, err := sessionService.Create(false, []string{}, "vanilla", 30)
		assert.NoError(t, err)

		session.Score = int32(30 - i*10)
		session.TotalAttempts = 10
		session.Accuracy = utils.CalculateAccuracy(session.Score, session.TotalAttempts)
		session.Phase = core.SessionPhaseFinished
		assert.NoError(t, sessionService.Update(session))
	}

	// Player's session — properly finished
	playerSession, err := sessionService.Create(false, []string{}, "vanilla", 30)
	assert.NoError(t, err)

	playerSession.Score = 5
	playerSession.TotalAttempts = 10
	playerSession.Accuracy = utils.CalculateAccuracy(5, 10)
	playerSession.Phase = core.SessionPhaseFinished // properly finished
	assert.NoError(t, sessionService.Update(playerSession))

	// Fetch leaderboard
	entries, err := leaderboardService.GetLeaderboard("vanilla", 10, 0)
	assert.NoError(t, err)

	total, err := leaderboardService.GetTotalEntries("vanilla")
	assert.NoError(t, err)

	// Correct: 4 entries
	assert.Len(t, entries, 4)
	assert.Equal(t, int64(4), total)
	// Player's entry should be last (worst score)
	assert.Equal(t, int32(5), entries[3].Score)
}

func BenchmarkLeaderboardService(b *testing.B) {
	db := testutils.SetupTestDB(b)
	service := NewLeaderboardService(db)

	sqlDB, _ := db.DB()
	txn, _ := sqlDB.Begin()
	stmt, _ := txn.Prepare(`INSERT INTO session_entities 
		(id, score, total_attempts, accuracy, correct_attempt_timestamps, used_words, word_definitions, used_item_ids, phase)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	var firstID string
	var firstScore int32
	for i := range 50_000 {
		id := utils.GenerateUUID()
		score := int32(rand.Intn(100))
		stmt.Exec(id, score, rand.Intn(100), rand.Float32()*100, "[]", "{}", "{}", "{}", core.SessionPhaseFinished)
		if i == 0 {
			firstID = id
			firstScore = score
		}
	}
	stmt.Close()
	txn.Commit()

	b.ResetTimer() // only times what's below

	for b.Loop() {
		start := time.Now()

		_, err := service.GetPercentile(firstID, "vanilla", firstScore)
		if err != nil {
			b.Fatal(err)
		}

		if elapsed := time.Since(start); elapsed > 10*time.Millisecond {
			b.Fatalf("too slow: %s, want < 100ms", elapsed)
		}
	}
}
