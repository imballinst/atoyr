package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"atoyr/server/internal/models"
	"atoyr/server/internal/utils"
)

type GameService struct {
	sessionService     *SessionService
	wordService        *WordService
	leaderboardService *LeaderboardService
}

type SubmitAnswerResult struct {
	Correct                 bool   `json:"correct"`
	ScrambledWord           string `json:"scrambledWord"`
	ScrambledWordDefinition string `json:"scrambledWordDefinition"`
	Token                   string `json:"token"`
	Score                   int32  `json:"score"`
	Attempts                int32  `json:"attempts"`
	RemainingSeconds        int32  `json:"remainingSeconds"`
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

	word, definition, err := g.wordService.GetRandomWord(session.UsedWords)
	if err != nil {
		// All words used, finish game
		return g.FinishGame(sessionID)
	}

	token := g.generateToken(word)
	if err := g.sessionService.SetCurrentWord(sessionID, word, definition, token); err != nil {
		return err
	}

	if err := g.sessionService.AddUsedWord(sessionID, word); err != nil {
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

	result := &SubmitAnswerResult{
		Correct:          false,
		Score:            session.Score,
		Attempts:         session.TotalAttempts + 1,
		RemainingSeconds: session.RemainingSeconds,
	}

	// Validate answer by checking token
	expectedToken := session.CurrentWordToken

	if token != expectedToken {
		log.Printf("submitted answer token: %s, expected %s\n", token, expectedToken)
	} else if answer != session.CurrentWord {
		log.Printf("submitted answer: %s, expected %s\n", answer, session.CurrentWord)
	} else {
		result.Correct = true
		if err := g.sessionService.UpdateScore(sessionID, 1); err != nil {
			return nil, err
		}

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
		result.Score = session.Score
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
