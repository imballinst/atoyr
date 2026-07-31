package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"atoyr/server/internal/core"
	"atoyr/server/internal/services/domainmodels"
	"atoyr/server/internal/utils"

	"gorm.io/gorm"
)

var (
	ErrSessionAlreadyFinished = fmt.Errorf("session is already finished")
	ErrSessionNotPlayingYet   = fmt.Errorf("session is not in %s phase yet", core.SessionPhasePlaying)
)

type GameService struct {
	sessionService     *SessionService
	wordService        *WordService
	leaderboardService *LeaderboardService
	sessionOptions     core.SessionOptions
}

type SubmitAnswerResult struct {
	Correct                  bool
	ScrambledWord            string
	ScrambledWordDefinition  string
	CorrectAttemptTimestamps [][]string
	Token                    string
	Score                    int32
	Attempts                 int32
	DurationSeconds          int32
}

func NewGameService(
	sessionService *SessionService,
	wordService *WordService,
	leaderboardService *LeaderboardService,
	sessionOptions core.SessionOptions,
) *GameService {
	return &GameService{
		sessionService:     sessionService,
		wordService:        wordService,
		leaderboardService: leaderboardService,
		sessionOptions:     sessionOptions,
	}
}

func (g *GameService) StartGame(sessionID string) (*domainmodels.SessionDomain, error) {
	// Start the game by emitting first word
	session, err := g.getSessionWithNextWord(sessionID)
	if err != nil {
		return nil, err
	}

	// Update phase to playing
	session.Phase = core.SessionPhasePlaying
	session.EndsAt = time.Now().Add(time.Second * time.Duration(session.DurationSeconds))

	if err := g.sessionService.Update(session); err != nil {
		return nil, err
	}

	// Start timer
	go g.startTimer(session)

	session, err = g.sessionService.FindByID(sessionID)
	if err != nil {
		return nil, err
	}

	session = g.sessionWithVisibleDefinition(session)
	return session, nil
}

func (g *GameService) ContinueGame(sessionID string) (*domainmodels.SessionDomain, error) {
	session, err := g.sessionService.FindByID(sessionID)
	if err == gorm.ErrRecordNotFound {
		return nil, gorm.ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}

	if session.Phase == core.SessionPhaseFinished {
		return nil, ErrSessionAlreadyFinished
	}
	if session.Phase != core.SessionPhasePlaying {
		return nil, ErrSessionNotPlayingYet
	}

	// Update seconds, because the second is paused prior to resume.
	if time.Now().Equal(session.EndsAt) || time.Now().After(session.EndsAt) {
		session.DurationSeconds = 0
		session.Phase = core.SessionPhaseFinished
	} else {
		session.DurationSeconds = int32(time.Until(session.EndsAt).Seconds())
	}

	err = g.sessionService.Update(session)
	if err != nil {
		return nil, fmt.Errorf("error when updating session, %+v", err.Error())
	}

	if session.Phase == core.SessionPhaseFinished {
		return nil, nil
	}

	session = g.sessionWithVisibleDefinition(session)

	// It would seem we can re-use the timer from the start game function.
	return session, err
}

func (g *GameService) SubmitAnswer(sessionID, answer, token string) (*SubmitAnswerResult, error) {
	session, err := g.sessionService.FindByID(sessionID)
	if err != nil {
		return nil, err
	}

	if session.Phase != core.SessionPhasePlaying {
		return nil, fmt.Errorf("game is not in playing state")
	}

	// Validate answer by checking token
	isTokenCorrect := token == session.CurrentWordToken
	isAnswerCorrect := answer == session.CurrentWord

	result := &SubmitAnswerResult{
		Correct:                  isTokenCorrect && isAnswerCorrect,
		Score:                    session.Score,
		Attempts:                 session.TotalAttempts + 1,
		DurationSeconds:          session.DurationSeconds,
		CorrectAttemptTimestamps: session.CorrectAttemptTimestamps,
		Token:                    session.CurrentWordToken,
	}

	if !isTokenCorrect {
		log.Printf("invalid token, submitted answer token: %s, expected %s\n", token, session.CurrentWordToken)
	} else if !isAnswerCorrect {
		log.Printf("invalid answer, submitted answer: %s, expected %s\n", answer, session.CurrentWord)
	} else {
		// No-op.
	}

	session, err = g.updateSessionBasedOnAnswerResult(session.ID, result.Correct)
	if err != nil {
		// If error is "all words used", finish game
		if err.Error() == "all words have been used" {
			g.FinishGame(sessionID)
			return result, nil
		}

		return nil, err
	}

	result.ScrambledWord = session.CurrentScrambledWord
	if session.Mode == "vanilla" {
		result.ScrambledWordDefinition = session.CurrentWordDefinition
	}
	result.CorrectAttemptTimestamps = session.CorrectAttemptTimestamps
	result.Token = session.CurrentWordToken
	result.Score = session.Score
	result.DurationSeconds = core.SessionDurationManager.Get(session.ID)

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

	session.Accuracy = utils.CalculateAccuracy(session.Score, session.TotalAttempts)
	session.EndsAt = time.Now()

	if err := g.sessionService.Update(session); err != nil {
		return err
	}

	core.SessionDurationManager.Clean(session.ID)

	return nil
}

func (g *GameService) startTimer(session *domainmodels.SessionDomain) {
	g.runTimer(session.ID, session.DurationSeconds)
}

func (g *GameService) RestoreTimer(sessionID string, remainingSeconds int32) {
	g.runTimer(sessionID, remainingSeconds)
}

func (g *GameService) runTimer(sessionID string, initial int32) {
	ticker := time.NewTicker(utils.ToDuration(g.sessionOptions.Tick))
	defer ticker.Stop()

	core.SessionDurationManager.Add(sessionID, initial)

	for range ticker.C {
		remainingSeconds := core.SessionDurationManager.Get(sessionID)
		if remainingSeconds <= 0 {
			if err := g.FinishGame(sessionID); err != nil {
				log.Println("error when finishing game due to time is 0, ", err.Error())
			}
			return
		}

		newRemaining := core.SessionDurationManager.Decrement(sessionID)
		if newRemaining <= 0 {
			if err := g.FinishGame(sessionID); err != nil {
				log.Println("error when finishing game due to time is 0, ", err.Error())
			}
			return
		}
	}
}

func (g *GameService) generateToken(word string) string {
	hash := sha256.Sum256([]byte(word))
	return hex.EncodeToString(hash[:])
}

func (g *GameService) getSessionWithNextWord(sessionID string) (*domainmodels.SessionDomain, error) {
	session, err := g.sessionService.FindByID(sessionID)
	if err != nil {
		return nil, err
	}

	word, definition, err := g.wordService.GetRandomWord(session.UsedWords, session.Topic)
	if err != nil {
		// All words used, finish game
		return nil, g.FinishGame(sessionID)
	}

	token := g.generateToken(word)
	newUsedWords := append(session.UsedWords, word)

	session.UsedWords = newUsedWords
	session.CurrentWordToken = token
	session.CurrentWordDefinition = definition
	session.CurrentWord = word

	return session, nil
}

func (g *GameService) updateSessionBasedOnAnswerResult(sessionID string, isCorrect bool) (*domainmodels.SessionDomain, error) {
	session, err := g.sessionService.FindByID(sessionID)
	if err != nil {
		return nil, err
	}

	session.TotalAttempts += 1

	if !isCorrect {
		core.SessionDurationManager.Decrement(session.ID)
		session.EndsAt = session.EndsAt.Add(-1 * time.Second)

		lastIdx := len(session.CorrectAttemptTimestamps) - 1
		if lastIdx >= 0 && len(session.CorrectAttemptTimestamps[lastIdx]) > 0 {
			session.CorrectAttemptTimestamps = append(session.CorrectAttemptTimestamps, []string{})
		}

		if err := g.sessionService.Update(session); err != nil {
			return nil, err
		}

		return session, nil
	}

	word, definition, err := g.wordService.GetRandomWord(session.UsedWords, session.Topic)
	if err != nil {
		// All words used, finish game
		return nil, g.FinishGame(sessionID)
	}

	token := g.generateToken(word)
	newUsedWords := append(session.UsedWords, word)

	session.UsedWords = newUsedWords
	session.CurrentWordToken = token
	session.CurrentWordDefinition = definition
	session.CurrentWord = word
	session.CurrentScrambledWord = utils.ScrambleWord(word)
	session.Score += 1

	if session.AutoVoice {
		core.SessionDurationManager.Extend(session.ID, core.BonusDurationPerWordWithAutoVoice)
		session.EndsAt = session.EndsAt.Add(time.Duration(core.BonusDurationPerWordWithAutoVoice) * time.Second)
	}

	if len(session.CorrectAttemptTimestamps) == 0 {
		session.CorrectAttemptTimestamps = append(session.CorrectAttemptTimestamps, []string{})
	}

	lastIdx := len(session.CorrectAttemptTimestamps) - 1

	currentStreakTimestamps := session.CorrectAttemptTimestamps[lastIdx]
	currentStreakTimestamps = append(currentStreakTimestamps, time.Now().Format(time.RFC3339))
	session.CorrectAttemptTimestamps[lastIdx] = currentStreakTimestamps

	if err := g.sessionService.Update(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (g *GameService) sessionWithVisibleDefinition(session *domainmodels.SessionDomain) *domainmodels.SessionDomain {
	if session.Mode == "blind" {
		session.CurrentWordDefinition = ""
	}
	return session
}
