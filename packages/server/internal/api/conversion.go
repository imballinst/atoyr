package api

import (
	"atoyr/server/internal/services"
	"atoyr/server/internal/utils"
)

func ToApiLeaderboardEntry(domainEntry services.LeaderboardEntry, sessionId string) LeaderboardEntry {
	isSessionSameAsCurrentUser := domainEntry.ID == sessionId

	return LeaderboardEntry{
		Id:                         utils.MaskSessionID(domainEntry.ID),
		IsSessionSameAsCurrentUser: &isSessionSameAsCurrentUser,
		Rank:                       domainEntry.Rank,
		Score:                      domainEntry.Score,
		TotalAttempts:              domainEntry.TotalAttempts,
		Accuracy:                   domainEntry.Accuracy,
		Timestamp:                  domainEntry.Timestamp,
		Topic:                      SessionTopic(domainEntry.Topic),
	}
}
