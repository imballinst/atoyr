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
	ErrSessionNotFound        = fmt.Errorf("session not found")
)

type GameService struct {
	sessionService     *SessionService
	sessionStore       *core.SessionStore
	wordService        *WordService
	leaderboardService *LeaderboardService
	agsSyncService     *AGSSyncService
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
	sessionStore *core.SessionStore,
	wordService *WordService,
	leaderboardService *LeaderboardService,
	agsSyncService *AGSSyncService,
	sessionOptions core.SessionOptions,
) *GameService {
	if sessionStore == nil {
		sessionStore = core.NewSessionStore()
	}
	return &GameService{
		sessionService:     sessionService,
		sessionStore:       sessionStore,
		wordService:        wordService,
		leaderboardService: leaderboardService,
		agsSyncService:     agsSyncService,
		sessionOptions:     sessionOptions,
	}
}

func (g *GameService) StartGame(sessionID, deviceID string, autoVoice bool) (*domainmodels.SessionDomain, error) {
	session, err := g.sessionService.FindByID(sessionID)
	if err != nil {
		return nil, err
	}

	if err := g.ensureAGSPlayer(session, deviceID); err != nil {
		return nil, err
	}

	word, definition, err := g.wordService.GetRandomWord(nil)
	if err != nil {
		return nil, err
	}

	token := g.generateToken(word)
	durationSeconds := g.sessionOptions.Duration
	if autoVoice {
		durationSeconds += int32(core.BonusDurationPerWordWithAutoVoice)
	}

	session.Phase = core.SessionPhasePlaying
	session.AutoVoice = autoVoice
	session.UsedItemIDs = []string{}
	session.UsedWords = []string{word}
	session.WordDefinitions = []string{definition}
	session.CurrentWord = word
	session.CurrentWordDefinition = definition
	session.CurrentScrambledWord = utils.ScrambleWord(word)
	session.CurrentWordToken = token
	session.Score = 0
	session.TotalAttempts = 0
	session.Accuracy = 0
	session.CorrectAttemptTimestamps = [][]string{}
	session.DurationSeconds = durationSeconds
	session.EndsAt = time.Now().Add(time.Second * time.Duration(durationSeconds))

	if err := g.sessionService.Update(session); err != nil {
		return nil, err
	}

	g.sessionStore.Set(session)
	core.SessionDurationManager.Add(session.ID, durationSeconds)
	go g.runTimer(session.ID, durationSeconds)

	g.syncRoundStateToCloudSave(session)
	g.sendSessionStarted(session)

	return g.sessionWithVisibleDefinition(session), nil
}

func (g *GameService) ContinueGame(sessionID string) (*domainmodels.SessionDomain, error) {
	session := g.sessionStore.Get(sessionID)
	if session != nil {
		if session.Phase == core.SessionPhaseFinished {
			return nil, ErrSessionAlreadyFinished
		}

		remaining := core.SessionDurationManager.Get(sessionID)
		if remaining <= 0 {
			session.Phase = core.SessionPhaseFinished
			session.DurationSeconds = 0
			g.FinishGame(sessionID)
			return nil, nil
		}

		session.DurationSeconds = remaining
		g.sessionStore.Set(session)
		return g.sessionWithVisibleDefinition(session), nil
	}

	// Fallback to the minimal registry. This path exists so that crash recovery
	// can still report session metadata even when the rich in-memory state has
	// not been restored from Cloud Save yet.
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

	remaining := int32(time.Until(session.EndsAt).Seconds())
	if remaining <= 0 {
		session.Phase = core.SessionPhaseFinished
		session.DurationSeconds = 0
		if err := g.sessionService.Update(session); err != nil {
			return nil, err
		}
		g.FinishGame(sessionID)
		return nil, nil
	}

	session.DurationSeconds = remaining
	return g.sessionWithVisibleDefinition(session), nil
}

func (g *GameService) SubmitAnswer(sessionID, answer, token string) (*SubmitAnswerResult, error) {
	session := g.sessionStore.Get(sessionID)
	if session == nil {
		return nil, ErrSessionNotFound
	}

	if session.Phase != core.SessionPhasePlaying {
		return nil, fmt.Errorf("game is not in playing state")
	}

	isTokenCorrect := token == session.CurrentWordToken
	isAnswerCorrect := answer == session.CurrentWord

	result := &SubmitAnswerResult{
		Correct:                  isTokenCorrect && isAnswerCorrect,
		Score:                    session.Score,
		Attempts:                 session.TotalAttempts + 1,
		DurationSeconds:          core.SessionDurationManager.Get(sessionID),
		CorrectAttemptTimestamps: session.CorrectAttemptTimestamps,
		Token:                    session.CurrentWordToken,
	}

	if !isTokenCorrect {
		log.Printf("invalid token, submitted answer token: %s, expected %s\n", token, session.CurrentWordToken)
	} else if !isAnswerCorrect {
		log.Printf("invalid answer, submitted answer: %s, expected %s\n", answer, session.CurrentWord)
	}

	updated, err := g.updateSessionBasedOnAnswerResult(session, result.Correct)
	if err != nil {
		if err.Error() == "all words have been used" {
			g.FinishGame(sessionID)
			return result, nil
		}
		return nil, err
	}

	result.ScrambledWord = updated.CurrentScrambledWord
	if updated.Mode == "vanilla" {
		result.ScrambledWordDefinition = updated.CurrentWordDefinition
	}
	result.CorrectAttemptTimestamps = updated.CorrectAttemptTimestamps
	result.Token = updated.CurrentWordToken
	result.Score = updated.Score
	result.DurationSeconds = core.SessionDurationManager.Get(sessionID)

	g.sessionStore.Set(updated)
	g.syncRoundStateToCloudSave(updated)

	return result, nil
}

func (g *GameService) FinishGame(sessionID string) error {
	session := g.sessionStore.Get(sessionID)
	if session == nil {
		registrySession, err := g.sessionService.FindByID(sessionID)
		if err != nil {
			return err
		}
		if registrySession.Phase == core.SessionPhaseFinished {
			return nil
		}
		session = registrySession
	}

	if err := g.sessionService.EndSession(sessionID); err != nil {
		return err
	}

	session.Phase = core.SessionPhaseFinished
	session.Accuracy = utils.CalculateAccuracy(session.Score, session.TotalAttempts)
	session.EndsAt = time.Now()

	if err := g.sessionService.Update(session); err != nil {
		return err
	}

	if g.leaderboardService != nil {
		if err := g.leaderboardService.UpsertEntry(session); err != nil {
			log.Printf("failed to upsert leaderboard entry: %v", err)
		}
	}

	g.postRoundStatsToAGS(session)
	g.sendSessionFinished(session)

	g.sessionStore.Delete(sessionID)
	core.SessionDurationManager.Clean(sessionID)

	return nil
}

func (g *GameService) startTimer(sessionID string, initial int32) {
	g.runTimer(sessionID, initial)
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

func (g *GameService) updateSessionBasedOnAnswerResult(session *domainmodels.SessionDomain, isCorrect bool) (*domainmodels.SessionDomain, error) {
	session.TotalAttempts += 1

	if !isCorrect {
		core.SessionDurationManager.Decrement(session.ID)
		session.EndsAt = session.EndsAt.Add(-1 * time.Second)

		lastIdx := len(session.CorrectAttemptTimestamps) - 1
		if lastIdx >= 0 && len(session.CorrectAttemptTimestamps[lastIdx]) > 0 {
			session.CorrectAttemptTimestamps = append(session.CorrectAttemptTimestamps, []string{})
		}

		return session, nil
	}

	word, definition, err := g.wordService.GetRandomWord(session.UsedWords)
	if err != nil {
		return nil, fmt.Errorf("all words have been used")
	}

	token := g.generateToken(word)
	session.UsedWords = append(session.UsedWords, word)
	session.WordDefinitions = append(session.WordDefinitions, definition)
	session.CurrentWord = word
	session.CurrentWordDefinition = definition
	session.CurrentWordToken = token
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

	return session, nil
}

func (g *GameService) sessionWithVisibleDefinition(session *domainmodels.SessionDomain) *domainmodels.SessionDomain {
	if session.Mode == "blind" {
		session.CurrentWordDefinition = ""
	}
	return session
}

func (g *GameService) ensureAGSPlayer(session *domainmodels.SessionDomain, deviceID string) error {
	if g.agsSyncService == nil || !g.agsSyncService.Enabled() {
		return nil
	}
	if session.UserID != "" {
		return nil
	}
	userID, _, err := g.agsSyncService.CreateHeadlessAccount(deviceID)
	if err != nil {
		return fmt.Errorf("failed to create AGS headless account: %w", err)
	}
	session.UserID = userID
	return g.sessionService.Update(session)
}

func (g *GameService) syncRoundStateToCloudSave(session *domainmodels.SessionDomain) {
	if g.agsSyncService == nil || !g.agsSyncService.Enabled() || session.UserID == "" {
		return
	}
	state := domainSessionToCloudSave(session)
	g.agsSyncService.SaveRoundStateAsync(session.UserID, session.ID, state)
}

func (g *GameService) postRoundStatsToAGS(session *domainmodels.SessionDomain) {
	if g.agsSyncService == nil || !g.agsSyncService.Enabled() || session.UserID == "" {
		return
	}
	durationSeconds := int32(time.Since(session.CreatedAt).Seconds())
	g.agsSyncService.PostRoundStatsAsync(session.UserID, session.Mode, session.Score, session.TotalAttempts, session.Accuracy, durationSeconds)
}

func (g *GameService) sendSessionStarted(session *domainmodels.SessionDomain) {
	if g.agsSyncService == nil || !g.agsSyncService.Enabled() || session.UserID == "" {
		return
	}
	go g.agsSyncService.SendSessionStarted(session.UserID, session.Mode, session.ID)
}

func (g *GameService) sendSessionFinished(session *domainmodels.SessionDomain) {
	if g.agsSyncService == nil || !g.agsSyncService.Enabled() || session.UserID == "" {
		return
	}
	durationSeconds := int32(time.Since(session.CreatedAt).Seconds())
	go g.agsSyncService.SendSessionFinished(session.UserID, session.Mode, session.ID, session.Score, session.TotalAttempts, session.Accuracy, durationSeconds)
}

func domainSessionToCloudSave(session *domainmodels.SessionDomain) CloudSaveRoundState {
	timestamps := make([]string, 0)
	for _, streak := range session.CorrectAttemptTimestamps {
		for _, ts := range streak {
			timestamps = append(timestamps, ts)
		}
	}
	return CloudSaveRoundState{
		Mode:                     session.Mode,
		Phase:                    session.Phase,
		AutoVoice:                session.AutoVoice,
		ItemsUsed:                session.UsedItemIDs,
		UsedWords:                session.UsedWords,
		CurrentWord:              session.CurrentWord,
		CurrentScrambledWord:     session.CurrentScrambledWord,
		CurrentWordDefinition:    session.CurrentWordDefinition,
		CurrentWordToken:         session.CurrentWordToken,
		Score:                    session.Score,
		TotalAttempts:            session.TotalAttempts,
		Accuracy:                 session.Accuracy,
		CorrectAttemptTimestamps: timestamps,
		EndsAt:                   session.EndsAt.UnixMilli(),
		CreatedAt:                session.CreatedAt.UnixMilli(),
		UpdatedAt:                time.Now().UnixMilli(),
	}
}
