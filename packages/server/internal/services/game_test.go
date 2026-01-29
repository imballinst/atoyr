package services

import (
	"testing"
)

func TestGameService_StartGame(t *testing.T) {
	db := setupTestDB(t)
	ws := setupTestWordService(t)
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	gs := NewGameService(ss, ws, ls)

	session, _ := ss.Create(false)
	started, err := gs.StartGame(session.ID)

	if err != nil {
		t.Fatalf("Failed to start game: %v", err)
	}

	if started.Phase != "playing" {
		t.Errorf("Expected phase playing, got %s", started.Phase)
	}

	if started.CurrentWord == "" {
		t.Error("Current word is empty after starting game")
	}

	if started.CurrentWordToken == "" {
		t.Error("Current word token is empty after starting game")
	}
}

func TestGameService_SubmitCorrectAnswer(t *testing.T) {
	db := setupTestDB(t)
	ws := setupTestWordService(t)
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	gs := NewGameService(ss, ws, ls)

	session, _ := ss.Create(false)
	gs.StartGame(session.ID)

	session, _ = ss.FindByID(session.ID)
	word := session.CurrentWord

	result, err := gs.SubmitAnswer(session.ID, word, session.CurrentWordToken)
	if err != nil {
		t.Fatalf("Failed to submit answer: %v", err)
	}

	if result.Correct {
		t.Error("Expected correct answer to be true")
	}

	if result.Score == 0 {
		t.Error("Expected score to be greater than 0")
	}
}

func TestGameService_SubmitIncorrectAnswer(t *testing.T) {
	db := setupTestDB(t)
	ws := setupTestWordService(t)
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	gs := NewGameService(ss, ws, ls)

	session, _ := ss.Create(false)
	gs.StartGame(session.ID)

	result, err := gs.SubmitAnswer(session.ID, "wronganswer", session.CurrentWordToken)
	if err != nil {
		t.Fatalf("Failed to submit answer: %v", err)
	}

	if !result.Correct {
		t.Error("Expected correct answer to be false")
	}
}

func TestGameService_TokenGeneration(t *testing.T) {
	db := setupTestDB(t)
	ws := setupTestWordService(t)
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	gs := NewGameService(ss, ws, ls)

	// Same word should generate same token
	token1 := gs.generateToken("test")
	token2 := gs.generateToken("test")

	if token1 != token2 {
		t.Error("Same word should generate same token")
	}

	// Different words should generate different tokens
	token3 := gs.generateToken("test2")
	if token1 == token3 {
		t.Error("Different words should generate different tokens")
	}
}

func TestGameService_AttemptCounting(t *testing.T) {
	db := setupTestDB(t)
	ws := setupTestWordService(t)
	ss := NewSessionService(db)
	ls := NewLeaderboardService(db)
	gs := NewGameService(ss, ws, ls)

	session, _ := ss.Create(false)
	gs.StartGame(session.ID)

	gs.SubmitAnswer(session.ID, "wronganswer", session.CurrentWordToken)
	session, _ = ss.FindByID(session.ID)
	if session.TotalAttempts != 1 {
		t.Errorf("Expected total attempts 1, got %d", session.TotalAttempts)
	}

	gs.SubmitAnswer(session.ID, "wronganswer2", session.CurrentWordToken)
	session, _ = ss.FindByID(session.ID)
	if session.TotalAttempts != 2 {
		t.Errorf("Expected total attempts 2, got %d", session.TotalAttempts)
	}
}
