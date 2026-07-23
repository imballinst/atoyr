package domainmodels

import (
	"atoyr/server/internal/models"
	"atoyr/server/internal/utils"
	"time"
)

// SessionDomain is the rich gameplay representation used by the game service.
// It is intentionally decoupled from the minimal SQLite registry model so that
// gameplay fields can be kept in memory while the registry only stores the
// columns needed for crash recovery and admin counting.
type SessionDomain struct {
	ID                       string
	UserID                   string
	CreatedAt                time.Time
	UpdatedAt                time.Time
	EndsAt                   time.Time
	Phase                    string
	Mode                     string
	Score                    int32
	TotalAttempts            int32
	Accuracy                 float32
	DurationSeconds          int32
	CorrectAttemptTimestamps [][]string
	AutoVoice                bool
	UsedWords                []string
	WordDefinitions          []string
	CurrentWord              string
	CurrentScrambledWord     string
	CurrentWordDefinition    string
	CurrentWordToken         string
	UsedItemIDs              []string
}

// ConvertSessionRegistryToDomain builds a SessionDomain from the minimal SQLite
// registry row. Gameplay fields are left at their zero values.
func ConvertSessionRegistryToDomain(session *models.SessionEntity) *SessionDomain {
	if session == nil {
		return nil
	}

	return &SessionDomain{
		ID:          session.ID,
		UserID:      session.UserID,
		CreatedAt:   session.CreatedAt,
		UpdatedAt:   session.UpdatedAt,
		EndsAt:      session.EndsAt,
		Phase:       session.Phase,
		Mode:        session.Mode,
		CurrentWord: session.CurrentWord,
	}
}

// ConvertSessionDomainToRegistry extracts only the registry fields from a rich
// SessionDomain. This is the only conversion used when writing to SQLite during
// gameplay.
func ConvertSessionDomainToRegistry(session *SessionDomain) *models.SessionEntity {
	if session == nil {
		return nil
	}

	return &models.SessionEntity{
		ID:          session.ID,
		UserID:      session.UserID,
		CreatedAt:   session.CreatedAt,
		UpdatedAt:   session.UpdatedAt,
		EndsAt:      session.EndsAt,
		Phase:       session.Phase,
		Mode:        session.Mode,
		CurrentWord: session.CurrentWord,
	}
}

// ConvertLeaderboardSessionToDomain builds a full SessionDomain from a
// LeaderboardSessionEntity. It is used when restoring or inspecting finished
// sessions from the SQLite leaderboard fallback.
func ConvertLeaderboardSessionToDomain(session *models.LeaderboardSessionEntity) (*SessionDomain, error) {
	if session == nil {
		return nil, nil
	}

	correctAttemptTimestamps, err := utils.ConvertDbJsonToNestedStringArray(session.CorrectAttemptTimestamps)
	if err != nil {
		return nil, err
	}

	return &SessionDomain{
		ID:                       session.ID,
		UserID:                   session.UserID,
		CreatedAt:                session.CreatedAt,
		UpdatedAt:                session.UpdatedAt,
		EndsAt:                   session.EndsAt,
		Phase:                    session.Phase,
		Mode:                     session.Mode,
		Score:                    session.Score,
		TotalAttempts:            session.TotalAttempts,
		Accuracy:                 session.Accuracy,
		DurationSeconds:          session.DurationSeconds,
		CorrectAttemptTimestamps: correctAttemptTimestamps,
		AutoVoice:                session.AutoVoice,
		UsedItemIDs:              []string(session.UsedItemIDs),
		UsedWords:                []string(session.UsedWords),
		WordDefinitions:          []string(session.WordDefinitions),
		CurrentWord:              session.CurrentWord,
		CurrentWordDefinition:    session.CurrentWordDefinition,
		CurrentScrambledWord:     session.CurrentScrambledWord,
		CurrentWordToken:         session.CurrentWordToken,
	}, nil
}

// ConvertSessionDomainToLeaderboardSession builds a LeaderboardSessionEntity
// from a rich SessionDomain. It is used to write finished session data to the
// SQLite leaderboard fallback.
func ConvertSessionDomainToLeaderboardSession(session *SessionDomain) (*models.LeaderboardSessionEntity, error) {
	if session == nil {
		return nil, nil
	}

	correctAttemptTimestamps, err := utils.ConvertNestedStringArrayToDbJson(session.CorrectAttemptTimestamps)
	if err != nil {
		return nil, err
	}

	return &models.LeaderboardSessionEntity{
		ID:                       session.ID,
		UserID:                   session.UserID,
		CreatedAt:                session.CreatedAt,
		UpdatedAt:                session.UpdatedAt,
		EndsAt:                   session.EndsAt,
		Phase:                    session.Phase,
		Mode:                     session.Mode,
		Score:                    session.Score,
		TotalAttempts:            session.TotalAttempts,
		Accuracy:                 session.Accuracy,
		DurationSeconds:          session.DurationSeconds,
		CorrectAttemptTimestamps: correctAttemptTimestamps,
		AutoVoice:                session.AutoVoice,
		UsedWords:                session.UsedWords,
		UsedItemIDs:              session.UsedItemIDs,
		WordDefinitions:          session.WordDefinitions,
		CurrentWord:              session.CurrentWord,
		CurrentScrambledWord:     session.CurrentScrambledWord,
		CurrentWordDefinition:    session.CurrentWordDefinition,
		CurrentWordToken:         session.CurrentWordToken,
	}, nil
}
