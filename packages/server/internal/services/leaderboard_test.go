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
			Topic:                    "english-words",
			Score:                    100,
			TotalAttempts:            10,
			Accuracy:                 90,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		},
		{
			ID:                       "session2",
			Mode:                     "vanilla",
			Topic:                    "english-words",
			Score:                    150,
			TotalAttempts:            15,
			Accuracy:                 95,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		},
		{
			ID:                       "session3",
			Mode:                     "vanilla",
			Topic:                    "english-words",
			Score:                    80,
			TotalAttempts:            8,
			Accuracy:                 85.12345,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		},
		{
			ID:                       "session4",
			Mode:                     "vanilla",
			Topic:                    "english-words",
			Score:                    0,
			TotalAttempts:            8,
			Accuracy:                 0,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		},
	}

	for _, result := range results {
		db.Create(&result)
	}

	// Get leaderboard
	entries, err := service.GetLeaderboard("vanilla", "english-words", 10, 0)
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
			Topic:                    "english-words",
			Score:                    100,
			TotalAttempts:            10,
			Accuracy:                 90,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		},
		{
			ID:                       "s02",
			Mode:                     "vanilla",
			Topic:                    "english-words",
			Score:                    100,
			TotalAttempts:            9,
			Accuracy:                 95,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		},
	}

	for _, result := range results {
		db.Create(&result)
	}

	// Get leaderboard
	entries, err := service.GetLeaderboard("vanilla", "english-words", 10, 0)
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
			Topic:                    "english-words",
			Score:                    100,
			TotalAttempts:            10,
			Accuracy:                 90,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		},
		{
			ID:                       "s02",
			Mode:                     "blind",
			Topic:                    "english-words",
			Score:                    100,
			TotalAttempts:            9,
			Accuracy:                 95,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		},
	}

	for _, result := range results {
		db.Create(&result)
	}

	// Get leaderboard
	entries, err := service.GetLeaderboard("vanilla", "english-words", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	assert.Equal(t, "s01", entries[0].ID)
	assert.Equal(t, float32(90), entries[0].Accuracy)

	// Get leaderboard, blind mode
	entries, err = service.GetLeaderboard("blind", "english-words", 10, 0)
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
			Topic:                    "english-words",
			Score:                    int32(100 + i),
			TotalAttempts:            int32(10),
			Accuracy:                 90,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		}
		db.Create(&result)
	}

	// Get first page (limit 10)
	entries, _ := service.GetLeaderboard("vanilla", "english-words", 10, 0)
	assert.Len(t, entries, 10)

	// Get second page
	entries, _ = service.GetLeaderboard("vanilla", "english-words", 10, 10)
	assert.Len(t, entries, 10)

	// Get third page (should have 5)
	entries, _ = service.GetLeaderboard("vanilla", "english-words", 10, 20)
	assert.Len(t, entries, 5)
}

func TestLeaderboardService_GetTotalEntries(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewLeaderboardService(db)

	total, _ := service.GetTotalEntries("vanilla", "english-words")
	assert.Equal(t, int64(0), total)

	// Add 5 results
	for i := range 5 {
		result := models.SessionEntity{
			ID:                       "session" + string(rune(i)),
			Mode:                     "vanilla",
			Topic:                    "english-words",
			Score:                    int32(100 + i),
			TotalAttempts:            10,
			Accuracy:                 90,
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
			UsedItemIDs:              pq.StringArray{},
			Phase:                    core.SessionPhaseFinished,
		}
		db.Create(&result)
	}

	total, _ = service.GetTotalEntries("vanilla", "english-words")
	assert.Equal(t, int64(5), total)
}

func TestLeaderboardService_GetLeaderboard_DifferentTopics(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewLeaderboardService(db)

	englishSession := models.SessionEntity{
		ID:                       "s01",
		Mode:                     "vanilla",
		Topic:                    "english-words",
		Score:                    100,
		TotalAttempts:            10,
		Accuracy:                 90,
		CorrectAttemptTimestamps: models.JSON{},
		UsedWords:                pq.StringArray{},
		UsedItemIDs:              pq.StringArray{},
		Phase:                    core.SessionPhaseFinished,
	}
	indonesianSession := models.SessionEntity{
		ID:                       "s02",
		Mode:                     "vanilla",
		Topic:                    "indonesian-politician-quotes",
		Score:                    200,
		TotalAttempts:            20,
		Accuracy:                 95,
		CorrectAttemptTimestamps: models.JSON{},
		UsedWords:                pq.StringArray{},
		UsedItemIDs:              pq.StringArray{},
		Phase:                    core.SessionPhaseFinished,
	}

	db.Create(&englishSession)
	db.Create(&indonesianSession)

	entries, err := service.GetLeaderboard("vanilla", "english-words", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "s01", entries[0].ID)

	entries, err = service.GetLeaderboard("vanilla", "indonesian-politician-quotes", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "s02", entries[0].ID)
}

func TestLeaderboardService_CrossTopic_GetTotalEntries(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewLeaderboardService(db)

	db.Create(&models.SessionEntity{
		ID:                       "s01",
		Mode:                     "vanilla",
		Topic:                    "english-words",
		Score:                    100,
		TotalAttempts:            10,
		Accuracy:                 90,
		CorrectAttemptTimestamps: models.JSON{},
		UsedWords:                pq.StringArray{},
		UsedItemIDs:              pq.StringArray{},
		Phase:                    core.SessionPhaseFinished,
	})

	total, err := service.GetTotalEntries("vanilla", "indonesian-politician-quotes")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)

	total, err = service.GetTotalEntries("vanilla", "english-words")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
}

func TestLeaderboardService_CrossTopic_Percentile(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewLeaderboardService(db)

	englishHigh := models.SessionEntity{
		ID:                       utils.GenerateUUID(),
		Mode:                     "vanilla",
		Topic:                    "english-words",
		Score:                    100,
		TotalAttempts:            10,
		Accuracy:                 90,
		CorrectAttemptTimestamps: models.JSON{},
		UsedWords:                pq.StringArray{},
		UsedItemIDs:              pq.StringArray{},
		Phase:                    core.SessionPhaseFinished,
	}
	englishMid := models.SessionEntity{
		ID:                       utils.GenerateUUID(),
		Mode:                     "vanilla",
		Topic:                    "english-words",
		Score:                    50,
		TotalAttempts:            10,
		Accuracy:                 80,
		CorrectAttemptTimestamps: models.JSON{},
		UsedWords:                pq.StringArray{},
		UsedItemIDs:              pq.StringArray{},
		Phase:                    core.SessionPhaseFinished,
	}
	englishLow := models.SessionEntity{
		ID:                       utils.GenerateUUID(),
		Mode:                     "vanilla",
		Topic:                    "english-words",
		Score:                    10,
		TotalAttempts:            10,
		Accuracy:                 70,
		CorrectAttemptTimestamps: models.JSON{},
		UsedWords:                pq.StringArray{},
		UsedItemIDs:              pq.StringArray{},
		Phase:                    core.SessionPhaseFinished,
	}
	indonesianSession := models.SessionEntity{
		ID:                       utils.GenerateUUID(),
		Mode:                     "vanilla",
		Topic:                    "indonesian-politician-quotes",
		Score:                    200,
		TotalAttempts:            10,
		Accuracy:                 100,
		CorrectAttemptTimestamps: models.JSON{},
		UsedWords:                pq.StringArray{},
		UsedItemIDs:              pq.StringArray{},
		Phase:                    core.SessionPhaseFinished,
	}

	db.Create(&englishHigh)
	db.Create(&englishMid)
	db.Create(&englishLow)
	db.Create(&indonesianSession)

	percentile, err := service.GetPercentile(englishMid.ID, "vanilla", "english-words", englishMid.Score)
	assert.NoError(t, err)
	assert.Equal(t, float32(50), percentile)
}

func TestLeaderboardService_GetPercentile(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewLeaderboardService(db)

	results := []models.SessionEntity{}

	for i := range 10 {
		result := models.SessionEntity{
			ID:                       utils.GenerateUUID(),
			Mode:                     "vanilla",
			Topic:                    "english-words",
			Score:                    int32(i * 10),
			TotalAttempts:            int32(10),
			Accuracy:                 float32(100 - (i * 10)),
			CorrectAttemptTimestamps: models.JSON{},
			UsedWords:                pq.StringArray{},
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

		percentile, err := service.GetPercentile(id, "vanilla", "english-words", score)
		assert.NoError(t, err)
		assert.Equal(t, expectedPercentile, percentile, map[string]any{"score": score, "totalBelowCurrentScore": totalBelowCurrentScore, "totalEligible": totalEligible})
	}

	// A blind session should only be compared against other blind sessions.
	percentile, err := service.GetPercentile(blindSession.ID, "blind", "english-words", blindSession.Score)
	assert.NoError(t, err)
	assert.Equal(t, float32(0), percentile)
}

// TestLeaderboardService_IncludesFinishedSession verifies that when FinishGame IS called
// (phase = "finished"), the player's session is included in the leaderboard.
func TestLeaderboardService_IncludesFinishedSession(t *testing.T) {
	db := testutils.SetupTestDB(t)
	leaderboardService := NewLeaderboardService(db)
	sessionService := NewSessionService(db)

	// Create 3 finished sessions with scores
	for i := range 3 {
		session, err := sessionService.Create(false, []string{}, "vanilla", testTopic, 30)
		assert.NoError(t, err)

		session.Score = int32(30 - i*10)
		session.TotalAttempts = 10
		session.Accuracy = utils.CalculateAccuracy(session.Score, session.TotalAttempts)
		session.Phase = core.SessionPhaseFinished
		assert.NoError(t, sessionService.Update(session))
	}

	// Player's session — properly finished
	playerSession, err := sessionService.Create(false, []string{}, "vanilla", testTopic, 30)
	assert.NoError(t, err)

	playerSession.Score = 5
	playerSession.TotalAttempts = 10
	playerSession.Accuracy = utils.CalculateAccuracy(5, 10)
	playerSession.Phase = core.SessionPhaseFinished // properly finished
	assert.NoError(t, sessionService.Update(playerSession))

	// Fetch leaderboard
	entries, err := leaderboardService.GetLeaderboard("vanilla", "english-words", 10, 0)
	assert.NoError(t, err)

	total, err := leaderboardService.GetTotalEntries("vanilla", "english-words")
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
		(id, mode, topic, score, total_attempts, accuracy, correct_attempt_timestamps, used_words, used_item_ids, phase)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	var firstID string
	var firstScore int32
	for i := range 50_000 {
		id := utils.GenerateUUID()
		score := int32(rand.Intn(100))
		stmt.Exec(id, "vanilla", "english-words", score, rand.Intn(100), rand.Float32()*100, "[]", "{}", "{}", core.SessionPhaseFinished)
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

		_, err := service.GetPercentile(firstID, "vanilla", "english-words", firstScore)
		if err != nil {
			b.Fatal(err)
		}

		if elapsed := time.Since(start); elapsed > 10*time.Millisecond {
			b.Fatalf("too slow: %s, want < 100ms", elapsed)
		}
	}
}
