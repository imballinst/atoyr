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

func ToApiLeaderboardEntry(domainEntry services.LeaderboardEntry) LeaderboardEntry {
	return LeaderboardEntry{
		Rank:          domainEntry.Rank,
		Score:         domainEntry.Score,
		TotalAttempts: domainEntry.TotalAttempts,
		Accuracy:      domainEntry.Accuracy,
		Timestamp:     domainEntry.Timestamp,
		User: &UserDomain{
			Id:       domainEntry.User.ID,
			Username: domainEntry.User.Username,
		},
	}
}
