package api

import (
	"atoyr/server/internal/core"
	"atoyr/server/internal/services"
	"atoyr/server/internal/utils"
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
		Id:                         utils.MaskSessionID(domainEntry.ID),
		IsSessionSameAsCurrentUser: &isSessionSameAsCurrentUser,
		Rank:                       domainEntry.Rank,
		Score:                      domainEntry.Score,
		TotalAttempts:              domainEntry.TotalAttempts,
		Accuracy:                   domainEntry.Accuracy,
		Timestamp:                  domainEntry.Timestamp,
		User:                       user,
	}
}
