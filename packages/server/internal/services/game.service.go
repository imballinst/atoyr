package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"atoyr/server/internal/models"
)

type GameService struct {
	sessionService     *SessionService
	wordService        *WordService
	leaderboardService *LeaderboardService
}

func NewGameService(sessionService *SessionService, wordService *WordService, leaderboardService *LeaderboardService) *GameService {
	return &GameService{
		sessionService:     sessionService,
		wordService:        wordService,
		leaderboardService: leaderboardService,
	}
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

	session, _ = g.sessionService.FindByID(sessionID)
	return session, nil
}

func (g *GameService) EmitWord(sessionID string) error {
	session, err := g.sessionService.FindByID(sessionID)
	if err != nil {
		return err
	}

	word, err := g.wordService.GetRandomWord(session.UsedWords)
	if err != nil {
		// All words used, finish game
		return g.FinishGame(sessionID)
	}

	token := g.generateToken(word)
	if err := g.sessionService.SetCurrentWord(sessionID, word, token); err != nil {
		return err
	}

	if err := g.sessionService.AddUsedWord(sessionID, word); err != nil {
		return err
	}

	return nil
}

func (g *GameService) SubmitAnswer(sessionID, answer string) (map[string]interface{}, error) {
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

	result := map[string]interface{}{
		"correct":   false,
		"score":     session.Score,
		"attempts":  session.TotalAttempts + 1,
		"remaining": session.RemainingSeconds,
	}

	// Validate answer by checking token
	expectedToken := g.generateToken(session.CurrentWord)
	answerToken := g.generateToken(answer)

	if answerToken == expectedToken {
		result["correct"] = true
		scoreIncrement := int32(session.RemainingSeconds)
		if err := g.sessionService.UpdateScore(sessionID, scoreIncrement); err != nil {
			return nil, err
		}
		result["score"] = session.Score + scoreIncrement

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
		result["newWord"] = session.CurrentWord
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
	accuracy := 0.0
	if session.TotalAttempts > 0 {
		accuracy = float64(session.Score/int32(session.TotalAttempts)) * 100
	}

	// Save result to leaderboard
	result := &models.ResultEntity{
		SessionID:     sessionID,
		Timestamp:     time.Now(),
		Score:         session.Score,
		TotalAttempts: session.TotalAttempts,
		Accuracy:      float32(accuracy),
		DurationMs:    (30 - session.RemainingSeconds) * 1000,
	}

	if err := g.sessionService.SaveResult(result); err != nil {
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
