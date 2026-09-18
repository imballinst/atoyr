package api

import (
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

	// Create 5 topics of mode vanilla. Upon migrating to v7, the old topics should be migrated.
	for range 5 {
		sessionService.Create(false, []string{}, string(Vanilla), string(SessionTopicEnglishWords), 30)
		sessionService.Create(false, []string{}, string(Vanilla), string(SessionTopicIndonesianPoliticianQuotes), 30)
	}

	englishWordsSessions := []models.SessionEntity{}
	indonesianPoliticianQuotesSessions := []models.SessionEntity{}

	err := db.Model(models.SessionEntity{}).Where("topic = ?", SessionTopicEnglishWords).Find(&englishWordsSessions).Error
	require.NoError(t, err)

	err = db.Model(models.SessionEntity{}).Where("topic = ?", SessionTopicIndonesianPoliticianQuotes).Find(&indonesianPoliticianQuotesSessions).Error
	require.NoError(t, err)

	assert.Len(t, englishWordsSessions, 5)
	assert.Len(t, indonesianPoliticianQuotesSessions, 5)

	err = testutils.RunMigrations(t, db, 7)
	require.NoError(t, err)

	oldEnglishWordsSessions := []models.SessionEntity{}
	englishWordsSessions = []models.SessionEntity{}
	indonesianPoliticianQuotesSessions = []models.SessionEntity{}

	err = db.Model(models.SessionEntity{}).Where("topic = ?", SessionTopicLeaderboardEnglishWordsJuly2026).Find(&oldEnglishWordsSessions).Error
	require.NoError(t, err)

	err = db.Model(models.SessionEntity{}).Where("topic = ?", SessionTopicEnglishWords).Find(&englishWordsSessions).Error
	require.NoError(t, err)

	err = db.Model(models.SessionEntity{}).Where("topic = ?", SessionTopicIndonesianPoliticianQuotes).Find(&indonesianPoliticianQuotesSessions).Error
	require.NoError(t, err)

	assert.Len(t, oldEnglishWordsSessions, 5)
	assert.Len(t, englishWordsSessions, 0)
	assert.Len(t, indonesianPoliticianQuotesSessions, 5)
}
