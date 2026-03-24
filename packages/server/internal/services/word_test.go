package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWordService_GetRandomWord(t *testing.T) {
	ws := setupTestWordService()

	word, _, err := ws.GetRandomWord([]string{})
	assert.NoError(t, err)
	assert.NotEqual(t, "", word)

	// Verify word is in the list
	found := false
	for _, w := range ws.words {
		if w.Word == word {
			found = true
			break
		}
	}

	assert.True(t, found, "Word %s not in word list", word)
}

func TestWordService_ExcludeWords(t *testing.T) {
	ws := setupTestWordService()

	// Exclude all words except one
	exclude := []string{"hello", "world", "apple", "banana", "cherry", "dragon", "elephant", "forest", "guitar"}
	word, _, err := ws.GetRandomWord(exclude)
	assert.NoError(t, err)
	assert.Equal(t, "horizon", word)
}

func TestWordService_AllWordsUsed(t *testing.T) {
	ws := setupTestWordService()

	// Exclude all words
	excluded := []string{}
	for _, w := range ws.words {
		excluded = append(excluded, w.Word)
	}

	_, _, err := ws.GetRandomWord(excluded)
	assert.Error(t, err)
	assert.Equal(t, "all words have been used", err.Error())
}

func TestWordService_CaseInsensitive(t *testing.T) {
	ws := setupTestWordService()

	// Exclude with different case
	Excluded := []string{"HELLO", "World", "APPLE"}
	word, _, err := ws.GetRandomWord(Excluded)
	assert.NoError(t, err)

	// Verify excluded words are not returned
	assert.NotEqual(t, "hello", word)
	assert.NotEqual(t, "world", word)
	assert.NotEqual(t, "apple", word)
}
