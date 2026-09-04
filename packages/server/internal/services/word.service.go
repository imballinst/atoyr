package services

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type WordService struct {
	words      map[string][]WordDefinition
	topicLangs map[string]string
	topics     []string
}

type TopicFile struct {
	Lang  string           `json:"lang"`
	Words []WordDefinition `json:"words"`
}

type WordDefinition struct {
	Word       string   `json:"word"`
	Definition string   `json:"definition"`
	References []string `json:"references"`
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

	w.words = make(map[string][]WordDefinition)
	w.topicLangs = make(map[string]string)

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

		var topicFile TopicFile
		if err := json.Unmarshal(data, &topicFile); err != nil {
			return fmt.Errorf("failed to parse topic file %s: %w", entry.Name(), err)
		}

		w.words[topic] = topicFile.Words
		w.topicLangs[topic] = topicFile.Lang
	}

	w.topics = make([]string, 0, len(w.words))
	for topic := range w.words {
		w.topics = append(w.topics, topic)
	}
	slices.Sort(w.topics)

	if len(w.words) == 0 {
		return fmt.Errorf("no topic files found in %s", topicsDir)
	}

	return nil
}

func (w *WordService) GetRandomWord(excludeWords []string, topic string) (*WordDefinition, error) {
	topicWords, ok := w.words[topic]
	if !ok {
		return nil, fmt.Errorf("unknown topic: %s", topic)
	}

	if len(topicWords) == 0 {
		return nil, fmt.Errorf("no words available for topic: %s", topic)
	}

	excluded := make(map[string]bool)
	for _, word := range excludeWords {
		excluded[strings.ToLower(word)] = true
	}

	available := []WordDefinition{}
	for _, wd := range topicWords {
		if !excluded[strings.ToLower(wd.Word)] {
			wd.Word = strings.ToLower(wd.Word)
			available = append(available, wd)
		}
	}

	if len(available) == 0 {
		return nil, fmt.Errorf("all words have been used")
	}

	idx := rand.Intn(len(available))
	return &available[idx], nil
}

func (w *WordService) GetTopicLang(topic string) string {
	return w.topicLangs[topic]
}

func (w *WordService) GetTopics() []string {
	return w.topics
}

func (w *WordService) SetWords(words map[string][]WordDefinition) {
	w.words = words
	w.topicLangs = map[string]string{}
	w.topics = []string{}

	for key := range words {
		w.topicLangs[key] = "en-US"
		w.topics = append(w.topics, key)
	}
}
