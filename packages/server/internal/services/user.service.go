package services

import (
	"fmt"
	"time"

	"atoyr/server/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	bonusTimerRewardItemID = "938d5ff9fda94b8098b02cc891093d10"
)

type UserService struct {
	db             *gorm.DB
	sessionService *SessionService
}

func NewUserService(sessionService *SessionService) *UserService {
	return &UserService{
		db:             sessionService.db,
		sessionService: sessionService,
	}
}

func (u *UserService) UpsertUserFromSession(username, sessionID string) (*models.UserEntity, error) {
	session, err := u.sessionService.FindByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	user, err := u.getOrCreateUser(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create user: %w", err)
	}

	session.UserEntity = *user

	err = u.sessionService.Update(session)
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	rewardItemIDs := []string{}
	for _, streak := range session.CorrectAttemptTimestamps {
		if streak > 3 {
			rewardItemIDs = append(rewardItemIDs, bonusTimerRewardItemID)
		}
	}

	return user, nil
}

func (u *UserService) getOrCreateUser(username string) (*models.UserEntity, error) {
	var user models.UserEntity
	err := u.db.First(&user, "username = ?", username).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new user
			user = models.UserEntity{
				ID:        uuid.New().String(),
				Username:  username,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := u.db.Create(&user).Error; err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			return &user, nil
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	return &user, nil
}
