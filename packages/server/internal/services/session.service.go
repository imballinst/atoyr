package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"atoyr/server/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type SessionService struct {
	db *gorm.DB
}

func NewSessionService(db *gorm.DB) *SessionService {
	return &SessionService{db: db}
}

func (s *SessionService) Create(autoVoice bool) (*models.SessionEntity, error) {
	session := &models.SessionEntity{
		ID:                       uuid.New().String(),
		Phase:                    "idle",
		Score:                    0,
		TotalAttempts:            0,
		RemainingSeconds:         30,
		AutoVoice:                autoVoice,
		UsedWords:                pq.StringArray{},
		WordDefinitions:          pq.StringArray{},
		CorrectAttemptTimestamps: models.JSON([]byte("[]")),
		CurrentWordDefinition:    "",
		CurrentWord:              "",
		CurrentWordToken:         "",
		CreatedAt:                time.Now(),
		ExpiresAt:                time.Now().Add(5 * time.Minute),
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func (s *SessionService) FindByID(id string) (*models.SessionEntity, error) {
	var session models.SessionEntity
	if err := s.db.First(&session, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}
	return &session, nil
}

func (s *SessionService) Update(session *models.SessionEntity) error {
	if err := s.db.Save(session).Error; err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}

func (s *SessionService) SetCurrentWord(sessionID, word, wordDefinition, token string, usedWords []string) error {
	if err := s.db.Model(&models.SessionEntity{}).
		Where("id = ?", sessionID).
		Updates(map[string]interface{}{
			"current_word":            word,
			"current_word_definition": wordDefinition,
			"current_word_token":      token,
			"used_words":              pq.StringArray(usedWords),
		}).Error; err != nil {
		return fmt.Errorf("failed to set current word: %w", err)
	}
	return nil
}

func (s *SessionService) UpdatePhase(sessionID, phase string) error {
	if err := s.db.Model(&models.SessionEntity{}).
		Where("id = ?", sessionID).
		Update("phase", phase).Error; err != nil {
		return fmt.Errorf("failed to update phase: %w", err)
	}
	return nil
}

func (s *SessionService) UpdateScore(sessionID string, scoreIncrement int32, correctAttemptTimestamps [][]string) error {
	correctAttemptTimestampsJson, err := json.Marshal(correctAttemptTimestamps)
	if err != nil {
		return fmt.Errorf("failed to marshal correct attempt timestamps: %w", err)
	}

	if err := s.db.Model(&models.SessionEntity{}).
		Where("id = ?", sessionID).
		Updates(map[string]interface{}{
			"score":                      gorm.Expr("score + ?", scoreIncrement),
			"correct_attempt_timestamps": correctAttemptTimestampsJson,
		}).Error; err != nil {
		return fmt.Errorf("failed to update score: %w", err)
	}
	return nil
}

func (s *SessionService) IncrementTotalAttempts(sessionID string) error {
	if err := s.db.Model(&models.SessionEntity{}).
		Where("id = ?", sessionID).
		Update("total_attempts", gorm.Expr("total_attempts + 1")).Error; err != nil {
		return fmt.Errorf("failed to increment total attempts: %w", err)
	}
	return nil
}

func (s *SessionService) SaveResult(result *models.ResultEntity) error {
	if err := s.db.Create(result).Error; err != nil {
		return fmt.Errorf("failed to save result: %w", err)
	}
	return nil
}

func (s *SessionService) UpdateRemainingSeconds(sessionID string, seconds int32) error {
	if err := s.db.Model(&models.SessionEntity{}).
		Where("id = ?", sessionID).
		Update("remaining_seconds", seconds).Error; err != nil {
		return fmt.Errorf("failed to update remaining seconds: %w", err)
	}
	return nil
}
