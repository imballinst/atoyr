package api

import (
	"atoyr/server/internal/services"
	"atoyr/server/internal/utils"
)

func ToApiLeaderboardEntry(domainEntry services.LeaderboardEntry, userID string) LeaderboardEntry {
	isSessionSameAsCurrentUser := domainEntry.ID == userID

	return LeaderboardEntry{
		Id:                         utils.MaskSessionID(domainEntry.ID),
		IsSessionSameAsCurrentUser: &isSessionSameAsCurrentUser,
		Rank:                       domainEntry.Rank,
		Score:                      domainEntry.Score,
		TotalAttempts:              domainEntry.TotalAttempts,
		Accuracy:                   domainEntry.Accuracy,
		Timestamp:                  domainEntry.Timestamp,
	}
}
