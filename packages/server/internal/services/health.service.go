package services

import (
	"net/http"
	"os"
	"sync"
	"time"

	"gorm.io/gorm"
)

type HealthService struct {
	name    string
	gitHash string
	db      *gorm.DB

	mu         sync.RWMutex
	databaseOK bool
	uiOK       bool
	cachedAt   time.Time
}

type HealthResponse struct {
	Name         string             `json:"name"`
	GitHash      string             `json:"gitHash"`
	Dependencies HealthDependencies `json:"dependencies"`
	CachedAt     time.Time          `json:"cachedAt"`
}

type HealthDependencies struct {
	Database bool `json:"database"`
	UI       bool `json:"ui"`
}

func NewHealthService(db *gorm.DB, name, gitHash string) *HealthService {
	hs := &HealthService{
		name:    name,
		gitHash: gitHash,
		db:      db,
	}

	go hs.refreshLoop()

	return hs
}

func (hs *HealthService) refreshLoop() {
	hs.refresh()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		hs.refresh()
	}
}

func (hs *HealthService) refresh() {
	databaseOK := false
	sqlDB, err := hs.db.DB()
	if err == nil {
		databaseOK = sqlDB.Ping() == nil
	}

	uiOK := true
	if os.Getenv("ENV") == "production" {
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("http://localhost:80/")
		if err == nil {
			resp.Body.Close()
			uiOK = resp.StatusCode < 500
		} else {
			uiOK = false
		}
	}

	hs.mu.Lock()
	hs.databaseOK = databaseOK
	hs.uiOK = uiOK
	hs.cachedAt = time.Now()
	hs.mu.Unlock()
}

func (hs *HealthService) GetHealth() *HealthResponse {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	return &HealthResponse{
		Name:    hs.name,
		GitHash: hs.gitHash,
		Dependencies: HealthDependencies{
			Database: hs.databaseOK,
			UI:       hs.uiOK,
		},
		CachedAt: hs.cachedAt,
	}
}
