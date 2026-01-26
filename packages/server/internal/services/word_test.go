package services

import (
	"testing"
)

func TestWordService_GetRandomWord(t *testing.T) {
	ws := setupTestWordService(t)

	word, err := ws.GetRandomWord([]string{})
	if err != nil {
		t.Fatalf("Failed to get random word: %v", err)
	}

	if word == "" {
		t.Error("Random word is empty")
	}

	// Verify word is in the list
	found := false
	for _, w := range ws.words {
		if w == word {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Word %s not in word list", word)
	}
}

func TestWordService_ExcludeWords(t *testing.T) {
	ws := setupTestWordService(t)

	// Exclude all words except one
	exclude := []string{"hello", "world", "apple", "banana", "cherry", "dragon", "elephant", "forest", "guitar"}
	word, err := ws.GetRandomWord(exclude)
	if err != nil {
		t.Fatalf("Failed to get random word: %v", err)
	}

	if word != "horizon" {
		t.Errorf("Expected horizon, got %s", word)
	}
}

func TestWordService_AllWordsUsed(t *testing.T) {
	ws := setupTestWordService(t)

	// Exclude all words
	exclude := ws.words

	_, err := ws.GetRandomWord(exclude)
	if err == nil {
		t.Error("Expected error when all words are excluded")
	}

	if err.Error() != "all words have been used" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestWordService_CaseInsensitive(t *testing.T) {
	ws := setupTestWordService(t)

	// Exclude with different case
	exclude := []string{"HELLO", "World", "APPLE"}
	word, err := ws.GetRandomWord(exclude)
	if err != nil {
		t.Fatalf("Failed to get random word: %v", err)
	}

	// Verify excluded words are not returned
	if word == "hello" || word == "world" || word == "apple" {
		t.Errorf("Excluded word returned: %s", word)
	}
}
