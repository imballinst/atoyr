package services

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
)

type WordService struct {
	Words []WordDefinition
}

type WordDefinition struct {
	Word       string `json:"word"`
	Definition string `json:"definition"`
}

func NewWordService() (*WordService, error) {
	ws := &WordService{}
	if err := ws.loadWords(); err != nil {
		return nil, err
	}
	return ws, nil
}

func (w *WordService) loadWords() error {
	wordsPath := os.Getenv("WORDS_PATH")
	if wordsPath == "" {
		return fmt.Errorf("WORDS_PATH environment variable is not set")
	}

	data, err := os.ReadFile(wordsPath)
	if err != nil {
		return fmt.Errorf("failed to read words file: %w", err)
	}

	var dbContent []WordDefinition

	if err := json.Unmarshal(data, &dbContent); err != nil {
		return fmt.Errorf("failed to parse words file: %w", err)
	}

	w.Words = dbContent
	return nil
}

func (w *WordService) GetRandomWord(excludeWords []string) (string, string, error) {
	if len(w.Words) == 0 {
		return "", "", fmt.Errorf("no words available")
	}

	// Create a map for excluded words for O(1) lookup
	excluded := make(map[string]bool)
	for _, word := range excludeWords {
		excluded[strings.ToLower(word)] = true
	}

	// Find available words
	available := []string{}
	availableDefinitions := []string{}
	for _, word := range w.Words {
		if !excluded[strings.ToLower(word.Word)] {
			available = append(available, word.Word)
			availableDefinitions = append(availableDefinitions, word.Definition)
		}
	}

	if len(available) == 0 {
		return "", "", fmt.Errorf("all words have been used")
	}

	idx := rand.Intn(len(available))
	return available[idx], availableDefinitions[idx], nil
}

func (w *WordService) SetWords(words []WordDefinition) {
	w.Words = words
}
