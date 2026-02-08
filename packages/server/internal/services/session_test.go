package services

import (
	"testing"
)

func TestSessionService_Create(t *testing.T) {
	db := setupTestDB(t)
	service := NewSessionService(db)

	session, err := service.Create(true)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.ID == "" {
		t.Error("Session ID is empty")
	}

	if session.Phase != "idle" {
		t.Errorf("Expected phase idle, got %s", session.Phase)
	}

	if session.Score != 0 {
		t.Errorf("Expected score 0, got %d", session.Score)
	}

	if session.RemainingSeconds != 30 {
		t.Errorf("Expected remaining seconds 30, got %d", session.RemainingSeconds)
	}

	if session.AutoVoice != true {
		t.Errorf("Expected autoVoice true, got %v", session.AutoVoice)
	}
}

func TestSessionService_FindByID(t *testing.T) {
	db := setupTestDB(t)
	service := NewSessionService(db)

	// Create session
	created, err := service.Create(false)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Find session
	found, err := service.FindByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to find session: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, found.ID)
	}
}

func TestSessionService_SetCurrentWord(t *testing.T) {
	db := setupTestDB(t)
	service := NewSessionService(db)

	session, _ := service.Create(false)

	err := service.SetCurrentWord(session.ID, "test", "definition", "token123", []string{"test"})
	if err != nil {
		t.Fatalf("Failed to set current word: %v", err)
	}

	found, _ := service.FindByID(session.ID)
	if found.CurrentWord != "test" {
		t.Errorf("Expected current word test, got %s", found.CurrentWord)
	}

	if found.CurrentWordToken != "token123" {
		t.Errorf("Expected token token123, got %s", found.CurrentWordToken)
	}

	if len(found.UsedWords) != 1 || found.UsedWords[0] != "test" {
		t.Errorf("Expected used words [test], got %v", found.UsedWords)
	}
}

func TestSessionService_UpdateScore(t *testing.T) {
	db := setupTestDB(t)
	service := NewSessionService(db)

	session, _ := service.Create(false)

	service.UpdateScore(session.ID, 10, [][]string{})
	found, _ := service.FindByID(session.ID)

	if found.Score != 10 {
		t.Errorf("Expected score 10, got %d", found.Score)
	}

	service.UpdateScore(session.ID, 5, [][]string{})
	found, _ = service.FindByID(session.ID)

	if found.Score != 15 {
		t.Errorf("Expected score 15, got %d", found.Score)
	}
}

func TestSessionService_IncrementTotalAttempts(t *testing.T) {
	db := setupTestDB(t)
	service := NewSessionService(db)

	session, _ := service.Create(false)

	service.IncrementTotalAttempts(session.ID)
	found, _ := service.FindByID(session.ID)

	if found.TotalAttempts != 1 {
		t.Errorf("Expected total attempts 1, got %d", found.TotalAttempts)
	}

	service.IncrementTotalAttempts(session.ID)
	found, _ = service.FindByID(session.ID)

	if found.TotalAttempts != 2 {
		t.Errorf("Expected total attempts 2, got %d", found.TotalAttempts)
	}
}

func TestSessionService_UpdatePhase(t *testing.T) {
	db := setupTestDB(t)
	service := NewSessionService(db)

	session, _ := service.Create(false)

	service.UpdatePhase(session.ID, "playing")
	found, _ := service.FindByID(session.ID)

	if found.Phase != "playing" {
		t.Errorf("Expected phase playing, got %s", found.Phase)
	}
}
