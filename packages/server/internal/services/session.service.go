package services

import (
	"fmt"
	"time"

	"atoyr/server/internal/core"
	"atoyr/server/internal/models"
	"atoyr/server/internal/services/domainmodels"
	"atoyr/server/internal/utils"

	"gorm.io/gorm"
)

type SessionService struct {
	db *gorm.DB
}

func NewSessionService(db *gorm.DB) *SessionService {
	return &SessionService{db: db}
}

// Create inserts a minimal registry row. The autoVoice and itemsUsed parameters
// are accepted for API compatibility but are not persisted in the registry;
// gameplay state lives in the in-memory SessionStore.
func (s *SessionService) Create(autoVoice bool, itemsUsed []string, mode string, durationSeconds int32) (*domainmodels.SessionDomain, error) {
	initialRemainingSeconds := durationSeconds
	if autoVoice {
		initialRemainingSeconds += 5
	}

	now := time.Now()
	session := &models.SessionEntity{
		ID:        utils.GenerateUUID(),
		UserID:    "",
		Phase:     "idle",
		Mode:      mode,
		CreatedAt: now,
		UpdatedAt: now.Add(5 * time.Minute),
		EndsAt:    now.Add(time.Duration(initialRemainingSeconds) * time.Second),
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	domain := domainmodels.ConvertSessionRegistryToDomain(session)
	// Preserve the values expected by callers for the returned domain, even
	// though they are not stored in the minimal registry.
	domain.AutoVoice = autoVoice
	domain.DurationSeconds = initialRemainingSeconds
	domain.UsedItemIDs = itemsUsed
	return domain, nil
}

func (s *SessionService) FindByID(id string) (*domainmodels.SessionDomain, error) {
	var session models.SessionEntity
	if err := s.db.First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return domainmodels.ConvertSessionRegistryToDomain(&session), nil
}

func (s *SessionService) Update(session *domainmodels.SessionDomain) error {
	dbModel := domainmodels.ConvertSessionDomainToRegistry(session)
	if dbModel == nil {
		return fmt.Errorf("failed to convert session to db model")
	}

	dbModel.UpdatedAt = time.Now()

	if err := s.db.Save(dbModel).Error; err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}

func (s *SessionService) EndSession(sessionID string) error {
	if err := s.db.Model(&models.SessionEntity{}).
		Where("id = ? AND phase != ?", sessionID, core.SessionPhaseFinished).
		Updates(map[string]any{
			"phase":      core.SessionPhaseFinished,
			"ends_at":    time.Now(),
			"updated_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to update phase: %w", err)
	}
	return nil
}

func (s *SessionService) FindPlaying() ([]*domainmodels.SessionDomain, error) {
	var entities []models.SessionEntity
	if err := s.db.Where("phase = ?", core.SessionPhasePlaying).Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find playing sessions: %w", err)
	}

	domains := make([]*domainmodels.SessionDomain, len(entities))
	for i, e := range entities {
		domains[i] = domainmodels.ConvertSessionRegistryToDomain(&e)
	}
	return domains, nil
}
