package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"atoyr/server/internal/core"
	"atoyr/server/internal/models"
	"atoyr/server/internal/services/domainmodels"
	"atoyr/server/internal/utils"

	"gorm.io/gorm"
)

var (
	ErrSessionNotPlayingYet = fmt.Errorf("session is not in %s phase yet", SessionPhasePlaying)
)

type GameService struct {
	sessionService     *SessionService
	wordService        *WordService
	leaderboardService *LeaderboardService

	items core.ItemInfoMap
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
	items core.ItemInfoMap,
) *GameService {

	return &GameService{
		sessionService:     sessionService,
		wordService:        wordService,
		leaderboardService: leaderboardService,
		items:              items,
	}
}

func (g *GameService) StartGame(sessionID string) (*models.SessionEntity, error) {
	// Start the game by emitting first word
	if err := g.EmitWord(sessionID); err != nil {
		return nil, err
	}

	session, err := g.sessionService.FindByID(sessionID)
	if err != nil {
		return nil, err
	}

	// Update phase to playing
	g.resolveItemEffects(session)
	session.Phase = SessionPhasePlaying
	session.EndsAt = time.Now().Add(time.Second * time.Duration(session.RemainingSeconds))

	if err := g.sessionService.Update(session); err != nil {
		return nil, err
	}

	// Start timer
	go g.startTimer(sessionID)

	session, err = g.sessionService.FindByID(sessionID)
	return session, err
}

func (g *GameService) ContinueGame(sessionID string) (*models.SessionEntity, error) {
	session, err := g.sessionService.FindByID(sessionID)
	if err == gorm.ErrRecordNotFound {
		return nil, gorm.ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}

	if session.Phase != SessionPhasePlaying {
		return nil, ErrSessionNotPlayingYet
	}

	// Update seconds, because the second is paused prior to resume.
	if time.Now().Equal(session.EndsAt) || time.Now().After(session.EndsAt) {
		session.RemainingSeconds = 0
		session.Phase = SessionPhaseFinished
	} else {
		fmt.Println(session.EndsAt, time.Until(session.EndsAt).Seconds())
		session.RemainingSeconds = int32(time.Until(session.EndsAt).Seconds())
	}

	err = g.sessionService.Update(session)
	if err != nil {
		return nil, fmt.Errorf("error when updating session, %+v", err.Error())
	}

	if session.Phase == SessionPhaseFinished {
		return nil, nil
	}

	// It would seem we can re-use the timer from the start game function.

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
	sessionEntity, err := g.sessionService.FindByID(sessionID)
	if err != nil {
		return nil, err
	}

	session, err := domainmodels.ConvertSessionDBToDomain(sessionEntity)
	if err != nil {
		return nil, err
	}

	if session.Phase != SessionPhasePlaying {
		return nil, fmt.Errorf("game is not in playing state")
	}

	// Validate answer by checking token
	result := &SubmitAnswerResult{
		Correct:                  false,
		Score:                    session.Score,
		Attempts:                 session.TotalAttempts + 1,
		RemainingSeconds:         session.RemainingSeconds,
		CorrectAttemptTimestamps: session.CorrectAttemptTimestamps,
	}

	scoreIncrement := int32(0)

	if token != session.CurrentWordToken {
		log.Printf("invalid token, submitted answer token: %s, expected %s\n", token, session.CurrentWordToken)

		lastIdx := len(session.CorrectAttemptTimestamps) - 1
		if lastIdx >= 0 {
			session.CorrectAttemptTimestamps = append(session.CorrectAttemptTimestamps, []string{})
		}
	} else if answer != session.CurrentWord {
		log.Printf("invalid answer, submitted answer: %s, expected %s\n", answer, session.CurrentWord)

		session.CorrectAttemptTimestamps = append(session.CorrectAttemptTimestamps, []string{})
		session.RemainingSeconds = session.RemainingSeconds - 1
	} else {
		if len(session.CorrectAttemptTimestamps) == 0 {
			session.CorrectAttemptTimestamps = append(session.CorrectAttemptTimestamps, []string{})
		}

		lastIdx := len(result.CorrectAttemptTimestamps) - 1
		currentStreakTimestamps := session.CorrectAttemptTimestamps[lastIdx]
		currentStreakTimestamps = append(currentStreakTimestamps, time.Now().Format(time.RFC3339))

		result.Correct = true
		session.CorrectAttemptTimestamps = append(session.CorrectAttemptTimestamps[:lastIdx], currentStreakTimestamps)
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
		sessionEntity, err := g.sessionService.FindByID(sessionID)
		if err != nil {
			return nil, err
		}

		session, err := domainmodels.ConvertSessionDBToDomain(sessionEntity)
		if err != nil {
			return nil, err
		}

		result.ScrambledWord = utils.ScrambleWord(session.CurrentWord)
		result.ScrambledWordDefinition = session.CurrentWordDefinition
		result.Token = g.generateToken(session.CurrentWord)
		result.Score += scoreIncrement
	}

	if err := g.sessionService.Update(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (g *GameService) FinishGame(sessionID string) error {
	if err := g.sessionService.EndSession(sessionID); err != nil {
		return err
	}

	session, err := g.sessionService.FindByID(sessionID)
	if err != nil {
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

		if session.Phase != SessionPhasePlaying {
			return
		}

		newRemaining := session.RemainingSeconds - 1
		if newRemaining <= 0 {
			err = g.FinishGame(sessionID)
			if err != nil {
				log.Println("error when finishing game due to time is 0, ", err.Error())
			}

			return
		}

		g.sessionService.UpdateRemainingSeconds(sessionID, newRemaining)
	}
}

func (g *GameService) generateToken(word string) string {
	hash := sha256.Sum256([]byte(word))
	return hex.EncodeToString(hash[:])
}

func (g *GameService) resolveItemEffects(session *models.SessionEntity) {
	// The assumption here is that the item IDs are already resolved in the inventory service.
	for _, itemID := range session.UsedItemIDs {
		if g.items[itemID].Kind == core.ItemTimerKind {
			session.RemainingSeconds += int32(g.items[itemID].Value)
		}
	}
}
