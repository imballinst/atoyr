package domainmodels

import (
	"atoyr/server/internal/models"
	"atoyr/server/internal/utils"
	"time"
)

type SessionDomain struct {
	ID                       string
	CreatedAt                time.Time
	UpdatedAt                time.Time
	EndsAt                   time.Time
	Phase                    string
	Score                    int32
	TotalAttempts            int32
	Accuracy                 float32
	RemainingSeconds         int32
	CorrectAttemptTimestamps [][]string
	AutoVoice                bool
	UsedWords                []string
	WordDefinitions          []string
	CurrentWord              string
	CurrentScrambledWord     string
	CurrentWordDefinition    string
	CurrentWordToken         string
	UsedItemIDs              []string
	UserEntity               *UserDomain
}

func ConvertSessionDBToDomain(session *models.SessionEntity) (*SessionDomain, error) {
	if session == nil {
		return nil, nil
	}

	correctAttemptTimestamps, err := utils.ConvertDbJsonToNestedStringArray(session.CorrectAttemptTimestamps)
	if err != nil {
		return nil, err
	}

	return &SessionDomain{
		ID:                       session.ID,
		CreatedAt:                session.CreatedAt,
		UpdatedAt:                session.UpdatedAt,
		EndsAt:                   session.EndsAt,
		Phase:                    session.Phase,
		Score:                    session.Score,
		TotalAttempts:            session.TotalAttempts,
		Accuracy:                 session.Accuracy,
		RemainingSeconds:         session.RemainingSeconds,
		CorrectAttemptTimestamps: correctAttemptTimestamps,
		AutoVoice:                session.AutoVoice,
		UsedItemIDs:              []string(session.UsedItemIDs),
		UsedWords:                []string(session.UsedWords),
		WordDefinitions:          []string(session.WordDefinitions),
		CurrentWord:              session.CurrentWord,
		CurrentWordDefinition:    session.CurrentWordDefinition,
		CurrentWordToken:         session.CurrentWordToken,
		UserEntity:               ConvertUserDBToDomain(session.UserEntity),
	}, nil
}

func ConvertSessionDomainToDB(session *SessionDomain) (*models.SessionEntity, error) {
	if session == nil {
		return nil, nil
	}

	correctAttemptTimestamps, err := utils.ConvertNestedStringArrayToDbJson(session.CorrectAttemptTimestamps)
	if err != nil {
		return nil, err
	}

	return &models.SessionEntity{
		ID:                       session.ID,
		CreatedAt:                session.CreatedAt,
		UpdatedAt:                session.UpdatedAt,
		EndsAt:                   session.EndsAt,
		Phase:                    session.Phase,
		Score:                    session.Score,
		TotalAttempts:            session.TotalAttempts,
		Accuracy:                 session.Accuracy,
		RemainingSeconds:         session.RemainingSeconds,
		CorrectAttemptTimestamps: correctAttemptTimestamps,
		AutoVoice:                session.AutoVoice,
		UsedWords:                session.UsedWords,
		UsedItemIDs:              session.UsedItemIDs,
		WordDefinitions:          session.WordDefinitions,
		CurrentWord:              session.CurrentWord,
		CurrentScrambledWord:     session.CurrentScrambledWord,
		CurrentWordDefinition:    session.CurrentWordDefinition,
		CurrentWordToken:         session.CurrentWordToken,
		UserEntity:               ConvertUserDomainToDB(session.UserEntity),
	}, nil
}

type UserDomain struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Username  string
}

func ConvertUserDBToDomain(user *models.UserEntity) *UserDomain {
	if user == nil {
		return nil
	}
	return &UserDomain{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Username:  user.Username,
	}
}

func ConvertUserDomainToDB(user *UserDomain) *models.UserEntity {
	if user == nil {
		return nil
	}
	return &models.UserEntity{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Username:  user.Username,
	}
}
