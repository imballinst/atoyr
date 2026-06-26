package services

import (
	"fmt"
	"time"

	"atoyr/server/internal/models"
	"atoyr/server/internal/services/domainmodels"
	"atoyr/server/internal/utils"

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

func (s *SessionService) Create(autoVoice bool, itemsUsed []string, durationSeconds int32) (*domainmodels.SessionDomain, error) {
	initialRemainingSeconds := durationSeconds
	if autoVoice {
		initialRemainingSeconds += 5
	}

	session := &models.SessionEntity{
		ID:                       utils.GenerateUUID(),
		Phase:                    "idle",
		Score:                    0,
		TotalAttempts:            0,
		DurationSeconds:          initialRemainingSeconds,
		AutoVoice:                autoVoice,
		UsedWords:                pq.StringArray{},
		WordDefinitions:          pq.StringArray{},
		CorrectAttemptTimestamps: models.JSON([]byte("[]")),
		CurrentWordDefinition:    "",
		CurrentWord:              "",
		CurrentWordToken:         "",
		UsedItemIDs:              itemsUsed,
		CreatedAt:                time.Now(),
		EndsAt:                   time.Now().Add(time.Duration(initialRemainingSeconds) * time.Second),
		UpdatedAt:                time.Now().Add(5 * time.Minute),
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return domainmodels.ConvertSessionDBToDomain(session)
}

func (s *SessionService) FindByID(id string) (*domainmodels.SessionDomain, error) {
	var session models.SessionEntity
	if err := s.db.First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return domainmodels.ConvertSessionDBToDomain(&session)
}

func (s *SessionService) Update(session *domainmodels.SessionDomain) error {
	dbModel, err := domainmodels.ConvertSessionDomainToDB(session)
	if err != nil {
		return fmt.Errorf("failed to convert session to db model: %w", err)
	}

	dbModel.UpdatedAt = time.Now()

	if err := s.db.Save(dbModel).Error; err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}

func (s *SessionService) EndSession(sessionID string) error {
	if err := s.db.Model(&models.SessionEntity{}).
		Where("id = ?", sessionID).
		Updates(map[string]any{
			"phase":      SessionPhaseFinished,
			"ends_at":    time.Now(),
			"updated_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to update phase: %w", err)
	}
	return nil
}
