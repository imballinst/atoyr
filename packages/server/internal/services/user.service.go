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

type UserService struct {
	db               *gorm.DB
	sessionService   *SessionService
	inventoryService *InventoryService
}

func NewUserService(sessionService *SessionService, inventoryService *InventoryService) *UserService {
	return &UserService{
		db:               sessionService.db,
		sessionService:   sessionService,
		inventoryService: inventoryService,
	}
}

func (u *UserService) UpsertUserFromSession(username, sessionID string) (*domainmodels.UserDomain, error) {
	session, err := u.sessionService.FindByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	user, err := u.getOrCreateUser(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create user: %w", err)
	}

	session.UserEntity = user

	err = u.sessionService.Update(session)
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	rewardItemIDs := []string{}
	for _, streak := range session.CorrectAttemptTimestamps {
		if len(streak) > 3 {
			rewardItemIDs = append(rewardItemIDs, core.BonusTimerRewardItemID)
		}
	}

	err = u.inventoryService.AddItems(rewardItemIDs, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to grant items from session reward: %w", err)
	}

	return user, nil
}

func (u *UserService) FindUserByID(id string) (*domainmodels.UserDomain, error) {
	var user models.UserEntity
	if err := u.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return domainmodels.ConvertUserDBToDomain(&user), nil
}

func (u *UserService) CreateUser(username string) (*domainmodels.UserDomain, error) {
	user := models.UserEntity{
		ID:        utils.GenerateUUID(),
		Username:  username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := u.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return domainmodels.ConvertUserDBToDomain(&user), nil
}

func (u *UserService) getOrCreateUser(username string) (*domainmodels.UserDomain, error) {
	var user models.UserEntity
	err := u.db.First(&user, "username = ?", username).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return u.CreateUser(username)
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	return domainmodels.ConvertUserDBToDomain(&user), nil
}
