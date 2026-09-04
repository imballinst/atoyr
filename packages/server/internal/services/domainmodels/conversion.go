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
	Mode                     string
	Topic                    string
	Score                    int32
	TotalAttempts            int32
	Accuracy                 float32
	DurationSeconds          int32
	CorrectAttemptTimestamps [][]string
	AutoVoice                bool
	UsedWords                []string
	CurrentWord              string
	CurrentScrambledWord     string
	CurrentWordDefinition    string
	CurrentWordToken         string
	CurrentWordOtherInfo     map[string]any
	UsedItemIDs              []string
}

func ConvertSessionDBToDomain(session *models.SessionEntity) (*SessionDomain, error) {
	if session == nil {
		return nil, nil
	}

	correctAttemptTimestamps, err := utils.ConvertDbJsonToNestedStringArray(session.CorrectAttemptTimestamps)
	if err != nil {
		return nil, err
	}

	otherInfo, err := utils.ConvertDbJsonStringToMap(session.CurrentWordOtherInfo)
	if err != nil {
		// No-op, it's already empty map anyway.
	}

	return &SessionDomain{
		ID:                       session.ID,
		CreatedAt:                session.CreatedAt,
		UpdatedAt:                session.UpdatedAt,
		EndsAt:                   session.EndsAt,
		Phase:                    session.Phase,
		Mode:                     session.Mode,
		Topic:                    session.Topic,
		Score:                    session.Score,
		TotalAttempts:            session.TotalAttempts,
		Accuracy:                 session.Accuracy,
		DurationSeconds:          session.DurationSeconds,
		CorrectAttemptTimestamps: correctAttemptTimestamps,
		AutoVoice:                session.AutoVoice,
		UsedItemIDs:              []string(session.UsedItemIDs),
		UsedWords:                []string(session.UsedWords),
		CurrentWord:              session.CurrentWord,
		CurrentWordDefinition:    session.CurrentWordDefinition,
		CurrentScrambledWord:     session.CurrentScrambledWord,
		CurrentWordToken:         session.CurrentWordToken,
		CurrentWordOtherInfo:     otherInfo,
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

	otherInfo, err := utils.ConvertMapToDbJson(session.CurrentWordOtherInfo)
	if err != nil {
		// No-op, it's already empty string anyway.
	}

	return &models.SessionEntity{
		ID:                       session.ID,
		CreatedAt:                session.CreatedAt,
		UpdatedAt:                session.UpdatedAt,
		EndsAt:                   session.EndsAt,
		Phase:                    session.Phase,
		Score:                    session.Score,
		Mode:                     session.Mode,
		Topic:                    session.Topic,
		TotalAttempts:            session.TotalAttempts,
		Accuracy:                 session.Accuracy,
		DurationSeconds:          session.DurationSeconds,
		CorrectAttemptTimestamps: correctAttemptTimestamps,
		AutoVoice:                session.AutoVoice,
		UsedWords:                session.UsedWords,
		UsedItemIDs:              session.UsedItemIDs,
		CurrentWord:              session.CurrentWord,
		CurrentScrambledWord:     session.CurrentScrambledWord,
		CurrentWordDefinition:    session.CurrentWordDefinition,
		CurrentWordToken:         session.CurrentWordToken,
		CurrentWordOtherInfo:     otherInfo,
	}, nil
}
