package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

type WordService struct {
	words []string
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
		// Default to the client data file
		wordsPath = filepath.Join(os.Getenv("PWD"), "..", "client", "src", "data", "words.json")
	}

	data, err := ioutil.ReadFile(wordsPath)
	if err != nil {
		return fmt.Errorf("failed to read words file: %w", err)
	}

	var response struct {
		Words []string `json:"words"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return fmt.Errorf("failed to parse words file: %w", err)
	}

	w.words = response.Words
	return nil
}

func (w *WordService) GetRandomWord(excludeWords []string) (string, error) {
	if len(w.words) == 0 {
		return "", fmt.Errorf("no words available")
	}

	// Create a map for excluded words for O(1) lookup
	excluded := make(map[string]bool)
	for _, word := range excludeWords {
		excluded[strings.ToLower(word)] = true
	}

	// Find available words
	available := []string{}
	for _, word := range w.words {
		if !excluded[strings.ToLower(word)] {
			available = append(available, word)
		}
	}

	if len(available) == 0 {
		return "", fmt.Errorf("all words have been used")
	}

	return available[rand.Intn(len(available))], nil
}
