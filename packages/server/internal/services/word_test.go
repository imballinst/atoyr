package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWordService_GetRandomWord(t *testing.T) {
	ws := initTestWordService()

	wordDef, err := ws.GetRandomWord([]string{}, "english-words")
	assert.NoError(t, err)
	assert.NotEqual(t, "", wordDef.Word)

	// Verify word is in the list
	found := false
	for _, w := range ws.words["english-words"] {
		if w.Word == wordDef.Word {
			found = true
			break
		}
	}

	assert.True(t, found, "Word %s not in word list", wordDef.Word)
}

func TestWordService_ExcludeWords(t *testing.T) {
	ws := initTestWordService()

	// Exclude all words except one
	exclude := []string{"hello", "world", "apple", "banana", "cherry", "dragon", "elephant", "forest", "guitar"}
	wordDef, err := ws.GetRandomWord(exclude, "english-words")
	assert.NoError(t, err)
	assert.Equal(t, "horizon", wordDef.Word)
}

func TestWordService_AllWordsUsed(t *testing.T) {
	ws := initTestWordService()

	// Exclude all words
	excluded := []string{}
	for _, w := range ws.words["english-words"] {
		excluded = append(excluded, w.Word)
	}

	_, err := ws.GetRandomWord(excluded, "english-words")
	assert.Error(t, err)
	assert.Equal(t, "all words have been used", err.Error())
}

func TestWordService_CaseInsensitive(t *testing.T) {
	ws := initTestWordService()

	// Exclude with different case
	Excluded := []string{"HELLO", "World", "APPLE"}
	wordDef, err := ws.GetRandomWord(Excluded, "english-words")
	assert.NoError(t, err)

	// Verify excluded words are not returned
	assert.NotEqual(t, "hello", wordDef.Word)
	assert.NotEqual(t, "world", wordDef.Word)
	assert.NotEqual(t, "apple", wordDef.Word)
}

func TestWordService_UnknownTopic(t *testing.T) {
	ws := initTestWordService()

	_, err := ws.GetRandomWord([]string{}, "unknown-topic")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown topic")
}

func TestWordService_TopicIsolation(t *testing.T) {
	ws := &WordService{
		words: map[string][]WordDefinition{
			"english-words": {
				{Word: "hello", Definition: "a greeting"},
				{Word: "world", Definition: "the earth"},
				{Word: "apple", Definition: "a fruit"},
				{Word: "banana", Definition: "a yellow fruit"},
				{Word: "cherry", Definition: "a red fruit"},
			},
			"indonesian-politician-quotes": {
				{Word: "gelap", Definition: "Kau yang <template>! — Pandjaitan, Luhut Binsar (2025)"},
				{Word: "internet", Definition: "<template> cepat buat apa? — Sembiring, Tifatul (2014)"},
				{Word: "obligasi", Definition: "<template> itu apa sih? — Widodo, Jokowi (2012)"},
				{Word: "rapat", Definition: "Gua ini kagak pernah diajakin <template> — Deyang, Nanik Sudaryati (2026)"},
				{Word: "pekerjaan", Definition: "Jika empat langkah tadi bisa penuhi akan terbuka 19 juta lapangan <template> — Rakabuming, Gibran (2023)"},
			},
		},
		topicLangs: map[string]string{
			"english-words":                "en-US",
			"indonesian-politician-quotes": "id-ID",
		},
	}

	for range 10 {
		wordDef, err := ws.GetRandomWord([]string{}, "english-words")
		assert.NoError(t, err)
		assert.Contains(t, []string{"hello", "world", "apple", "banana", "cherry"}, wordDef.Word)
	}

	for range 10 {
		wordDef, err := ws.GetRandomWord([]string{}, "indonesian-politician-quotes")
		assert.NoError(t, err)
		assert.Contains(t, []string{"gelap", "internet", "obligasi", "rapat", "pekerjaan"}, wordDef.Word)
	}
}
