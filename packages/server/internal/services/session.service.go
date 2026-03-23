package services

import (
	"fmt"
	"time"

	"atoyr/server/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

const (
	SessionPhasePlaying  = "playing"
	SessionPhaseFinished = "finished"
)

type SessionService struct {
	db *gorm.DB
}

func NewSessionService(db *gorm.DB) *SessionService {
	return &SessionService{db: db}
}

func (s *SessionService) Create(autoVoice bool, itemsUsed []string) (*models.SessionEntity, error) {
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
		UsedItemIDs:              pq.StringArray(itemsUsed),
		CreatedAt:                time.Now(),
		EndsAt:                   time.Now().Add(30 * time.Second),
		UpdatedAt:                time.Now().Add(5 * time.Minute),
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func (s *SessionService) FindByID(id string) (*models.SessionEntity, error) {
	var session models.SessionEntity
	if err := s.db.First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *SessionService) Update(session *models.SessionEntity) error {
	session.UpdatedAt = time.Now()

	if err := s.db.Save(session).Error; err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}

func (s *SessionService) EndSession(sessionID string) error {
	if err := s.db.Model(&models.SessionEntity{}).
		Where("id = ?", sessionID).
		Updates(map[string]any{
			"phase":             "finished",
			"remaining_seconds": 0,
		}).Error; err != nil {
		return fmt.Errorf("failed to update phase: %w", err)
	}
	return nil
}
