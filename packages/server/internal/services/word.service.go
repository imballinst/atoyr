package services

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

type WordService struct {
	Words map[string][]WordDefinition
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
	topicsDir := os.Getenv("TOPICS_DIR")
	if topicsDir == "" {
		return fmt.Errorf("TOPICS_DIR environment variable is not set")
	}

	entries, err := os.ReadDir(topicsDir)
	if err != nil {
		return fmt.Errorf("failed to read topics directory: %w", err)
	}

	w.Words = make(map[string][]WordDefinition)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		topic := strings.TrimSuffix(entry.Name(), ".json")
		filePath := filepath.Join(topicsDir, entry.Name())

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read topic file %s: %w", entry.Name(), err)
		}

		var entries []WordDefinition
		if err := json.Unmarshal(data, &entries); err != nil {
			return fmt.Errorf("failed to parse topic file %s: %w", entry.Name(), err)
		}

		w.Words[topic] = entries
	}

	if len(w.Words) == 0 {
		return fmt.Errorf("no topic files found in %s", topicsDir)
	}

	return nil
}

func (w *WordService) GetRandomWord(excludeWords []string, topic string) (string, string, error) {
	topicWords, ok := w.Words[topic]
	if !ok {
		return "", "", fmt.Errorf("unknown topic: %s", topic)
	}

	if len(topicWords) == 0 {
		return "", "", fmt.Errorf("no words available for topic: %s", topic)
	}

	excluded := make(map[string]bool)
	for _, word := range excludeWords {
		excluded[strings.ToLower(word)] = true
	}

	available := []string{}
	availableDefinitions := []string{}
	for _, wd := range topicWords {
		if !excluded[strings.ToLower(wd.Word)] {
			available = append(available, wd.Word)
			availableDefinitions = append(availableDefinitions, wd.Definition)
		}
	}

	if len(available) == 0 {
		return "", "", fmt.Errorf("all words have been used")
	}

	idx := rand.Intn(len(available))
	return available[idx], availableDefinitions[idx], nil
}

func (w *WordService) SetWords(words []WordDefinition) {
	w.Words = map[string][]WordDefinition{
		"english-words": words,
	}
}
