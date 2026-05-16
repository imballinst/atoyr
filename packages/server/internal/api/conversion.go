package api

import (
	"atoyr/server/internal/core"
	"atoyr/server/internal/services"
)

func toApiInventoryItem(domainItem core.InventoryItem) InventoryItem {
	return InventoryItem{
		Id:          domainItem.ItemID,
		Name:        domainItem.Name,
		Description: domainItem.Description,
		Quantity:    domainItem.Quantity,
		CreatedAt:   domainItem.CreatedAt,
		UpdatedAt:   domainItem.UpdatedAt,
	}
}

func ToApiLeaderboardEntry(domainEntry services.LeaderboardEntry, userId, sessionId string) LeaderboardEntry {
	var user *UserDomain
	isSessionSameAsCurrentUser := domainEntry.ID == sessionId

	if domainEntry.User != nil {
		user = &UserDomain{
			Id:       domainEntry.User.ID,
			Username: domainEntry.User.Username,
		}

		isSessionSameAsCurrentUser = isSessionSameAsCurrentUser || user.Id == userId
	}

	return LeaderboardEntry{
		Id:                         maskString(domainEntry.ID),
		IsSessionSameAsCurrentUser: isSessionSameAsCurrentUser,
		Rank:                       domainEntry.Rank,
		Score:                      domainEntry.Score,
		TotalAttempts:              domainEntry.TotalAttempts,
		Accuracy:                   domainEntry.Accuracy,
		Timestamp:                  domainEntry.Timestamp,
		User:                       user,
	}
}

func maskString(s string) string {
	rs := []rune(s)
	// Leave the last 4 characters unmasked
	for i := 0; i < len(rs)-4; i++ {
		rs[i] = '*'
	}
	return string(rs)
}
