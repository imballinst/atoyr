package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"atoyr/server/internal/models"
	"atoyr/server/internal/utils"
)

type item struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type GameService struct {
	sessionService     *SessionService
	wordService        *WordService
	leaderboardService *LeaderboardService

	items map[string]item
}

type SubmitAnswerResult struct {
	Correct                  bool       `json:"correct"`
	ScrambledWord            string     `json:"scrambledWord"`
	ScrambledWordDefinition  string     `json:"scrambledWordDefinition"`
	CorrectAttemptTimestamps [][]string `json:"correctAttemptTimestamps"`
	Token                    string     `json:"token"`
	Score                    int32      `json:"score"`
	Attempts                 int32      `json:"attempts"`
	RemainingSeconds         int32      `json:"remainingSeconds"`
}

func NewGameService(
	sessionService *SessionService,
	wordService *WordService,
	leaderboardService *LeaderboardService,
) (*GameService, error) {
	items, err := loadItems()
	if err != nil {
		return nil, err
	}

	return &GameService{
		sessionService:     sessionService,
		wordService:        wordService,
		leaderboardService: leaderboardService,
		items:              items,
	}, nil
}

func (g *GameService) StartGame(sessionID string) (*models.SessionEntity, error) {
	session, err := g.sessionService.FindByID(sessionID)
	if err != nil {
		return nil, err
	}

	// Start the game by emitting first word
	if err := g.EmitWord(sessionID); err != nil {
		return nil, err
	}

	// Update phase to playing
	if err := g.sessionService.UpdatePhase(sessionID, "playing"); err != nil {
		return nil, err
	}

	// Start timer
	go g.startTimer(sessionID)

	session, err = g.sessionService.FindByID(sessionID)
	return session, err
}

func (g *GameService) EmitWord(sessionID string) error {
	session, err := g.sessionService.FindByID(sessionID)
	if err != nil {
		return err
	}

	word, definition, err := g.wordService.GetRandomWord(session.UsedWords)
	if err != nil {
		// All words used, finish game
		return g.FinishGame(sessionID)
	}

	token := g.generateToken(word)
	newUsedWords := append(session.UsedWords, word)

	if err := g.sessionService.SetCurrentWord(sessionID, word, definition, token, newUsedWords); err != nil {
		return err
	}

	return nil
}

func (g *GameService) SubmitAnswer(sessionID, answer, token string) (*SubmitAnswerResult, error) {
	session, err := g.sessionService.FindByID(sessionID)
	if err != nil {
		return nil, err
	}

	if session.Phase != "playing" {
		return nil, fmt.Errorf("game is not in playing state")
	}

	if err := g.sessionService.IncrementTotalAttempts(sessionID); err != nil {
		return nil, err
	}

	correctAttemptTimestamps, err := convertTimestampJSONToStringArray(session.CorrectAttemptTimestamps)
	if err != nil {
		return nil, err
	}

	result := &SubmitAnswerResult{
		Correct:                  false,
		Score:                    session.Score,
		Attempts:                 session.TotalAttempts + 1,
		RemainingSeconds:         session.RemainingSeconds,
		CorrectAttemptTimestamps: correctAttemptTimestamps,
	}

	// Validate answer by checking token
	expectedToken := session.CurrentWordToken
	scoreIncrement := int32(0)

	if token != expectedToken {
		log.Printf("invalid token, submitted answer token: %s, expected %s\n", token, expectedToken)

		result.CorrectAttemptTimestamps = append(correctAttemptTimestamps, []string{})
	} else if answer != session.CurrentWord {
		log.Printf("invalid answer, submitted answer: %s, expected %s\n", answer, session.CurrentWord)

		result.CorrectAttemptTimestamps = append(correctAttemptTimestamps, []string{})
	} else {
		if len(result.CorrectAttemptTimestamps) == 0 {
			result.CorrectAttemptTimestamps = append(result.CorrectAttemptTimestamps, []string{})
		}

		lastIdx := len(result.CorrectAttemptTimestamps) - 1
		currentStreakTimestamps := result.CorrectAttemptTimestamps[lastIdx]
		currentStreakTimestamps = append(currentStreakTimestamps, time.Now().Format(time.RFC3339))

		result.Correct = true
		result.CorrectAttemptTimestamps = append(correctAttemptTimestamps[:lastIdx], currentStreakTimestamps)
		scoreIncrement = 1

		// Emit next word
		if err := g.EmitWord(sessionID); err != nil {
			// If error is "all words used", finish game
			if err.Error() == "all words have been used" {
				g.FinishGame(sessionID)
			}
			return result, nil
		}

		// Reset timer for next word
		session, _ = g.sessionService.FindByID(sessionID)
		result.ScrambledWord = utils.ScrambleWord(session.CurrentWord)
		result.ScrambledWordDefinition = session.CurrentWordDefinition
		result.Token = g.generateToken(session.CurrentWord)
		result.Score += scoreIncrement
	}

	if err := g.sessionService.UpdateScore(sessionID, scoreIncrement, result.CorrectAttemptTimestamps); err != nil {
		return nil, err
	}

	return result, nil
}

func (g *GameService) FinishGame(sessionID string) error {
	session, err := g.sessionService.FindByID(sessionID)
	if err != nil {
		return err
	}

	if err := g.sessionService.UpdatePhase(sessionID, "finished"); err != nil {
		return err
	}

	// Calculate accuracy
	accuracy := float32(0)
	if session.TotalAttempts > 0 {
		accuracy = float32(session.Score/int32(session.TotalAttempts)) * 100
	}
	session.Accuracy = accuracy
	session.EndsAt = time.Now()

	if err := g.sessionService.Update(session); err != nil {
		return err
	}

	return nil
}

func (g *GameService) startTimer(sessionID string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		session, err := g.sessionService.FindByID(sessionID)
		if err != nil {
			return
		}

		if session.Phase != "playing" {
			return
		}

		newRemaining := session.RemainingSeconds - 1
		if newRemaining <= 0 {
			g.FinishGame(sessionID)
			return
		}

		g.sessionService.UpdateRemainingSeconds(sessionID, newRemaining)
	}
}

func (g *GameService) generateToken(word string) string {
	hash := sha256.Sum256([]byte(word))
	return hex.EncodeToString(hash[:])
}

func (g *GameService) resolveItemEffects(session *models.SessionEntity, itemIDs []string) {
	for _, itemID := range itemIDs {
		if g.items[itemID].Kind == "timer" {
			session.RemainingSeconds += int32(g.items[itemID].Value)
		}
	}
}

func convertTimestampJSONToStringArray(j models.JSON) ([][]string, error) {
	var result [][]string
	err := json.Unmarshal([]byte(j), &result)
	return result, err
}

func loadItems() (map[string]item, error) {
	itemsPath := os.Getenv("ITEMS_PATH")
	if itemsPath == "" {
		return nil, fmt.Errorf("ITEMS_PATH environment variable is not set")
	}

	data, err := os.ReadFile(itemsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read items file: %w", err)
	}

	var dbContent []item

	if err := json.Unmarshal(data, &dbContent); err != nil {
		return nil, fmt.Errorf("failed to parse items file: %w", err)
	}

	itemRecord := map[string]item{}
	for _, item := range dbContent {
		itemRecord[item.ID] = item
	}

	return itemRecord, nil
}
