package api

import (
	"atoyr/server/internal/core"
	"atoyr/server/internal/models"
	"atoyr/server/internal/services"
	"atoyr/server/internal/testutils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration67_MigrateEnglishWordsTopic(t *testing.T) {
	db := testutils.SetupTestDBWithVersion(t, 6)

	sessionService := services.NewSessionService(db)

	// Finished english-words sessions form the old leaderboard and must move to
	// the immortalized topic. Playing sessions must keep their topic so they can
	// still be assigned words after the migration.
	for range 5 {
		finished, err := sessionService.Create(false, []string{}, string(Vanilla), string(SessionTopicEnglishWords), 30)
		require.NoError(t, err)
		require.NoError(t, sessionService.EndSession(finished.ID))

		playing, err := sessionService.Create(false, []string{}, string(Vanilla), string(SessionTopicEnglishWords), 30)
		require.NoError(t, err)
		require.NoError(t, db.Model(&models.SessionEntity{}).Where("id = ?", playing.ID).Update("phase", core.SessionPhasePlaying).Error)

		_, err = sessionService.Create(false, []string{}, string(Vanilla), string(SessionTopicIndonesianPoliticianQuotes), 30)
		require.NoError(t, err)
	}

	require.NoError(t, testutils.RunMigrations(t, db, 7))

	oldEnglishWordsSessions := []models.SessionEntity{}
	playingEnglishWordsSessions := []models.SessionEntity{}
	indonesianPoliticianQuotesSessions := []models.SessionEntity{}

	err := db.Model(models.SessionEntity{}).Where("topic = ?", SessionTopicLeaderboardEnglishWordsJuly2026).Find(&oldEnglishWordsSessions).Error
	require.NoError(t, err)

	err = db.Model(models.SessionEntity{}).
		Where("topic = ? AND phase = ?", SessionTopicEnglishWords, core.SessionPhasePlaying).
		Find(&playingEnglishWordsSessions).Error
	require.NoError(t, err)

	err = db.Model(models.SessionEntity{}).Where("topic = ?", SessionTopicIndonesianPoliticianQuotes).Find(&indonesianPoliticianQuotesSessions).Error
	require.NoError(t, err)

	assert.Len(t, oldEnglishWordsSessions, 5)
	assert.Len(t, playingEnglishWordsSessions, 5)
	assert.Len(t, indonesianPoliticianQuotesSessions, 5)
}
