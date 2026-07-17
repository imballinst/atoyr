package services

import (
	"sync"
	"time"

	"gorm.io/gorm"
)

type HealthService struct {
	name               string
	gitHash            string
	db                 *gorm.DB
	containerStartedAt time.Time

	mu         sync.RWMutex
	databaseOK bool
	cachedAt   time.Time
}

type HealthResponse struct {
	Name               string             `json:"name"`
	GitHash            string             `json:"gitHash"`
	ContainerStartedAt time.Time          `json:"containerStartedAt"`
	Dependencies       HealthDependencies `json:"dependencies"`
	CachedAt           time.Time          `json:"cachedAt"`
}

type HealthDependencies struct {
	Database bool `json:"database"`
}

func NewHealthService(db *gorm.DB, name, gitHash string) *HealthService {
	hs := &HealthService{
		name:               name,
		gitHash:            gitHash,
		db:                 db,
		containerStartedAt: time.Now().UTC(),
	}

	hs.refresh()
	go hs.refreshLoop()

	return hs
}

func (hs *HealthService) refreshLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		hs.refresh()
	}
}

func (hs *HealthService) refresh() {
	databaseOK := false
	if hs.db != nil {
		sqlDB, err := hs.db.DB()
		if err == nil {
			databaseOK = sqlDB.Ping() == nil
		}
	}

	hs.mu.Lock()
	hs.databaseOK = databaseOK
	hs.cachedAt = time.Now()
	hs.mu.Unlock()
}

func (hs *HealthService) GetHealth() *HealthResponse {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	return &HealthResponse{
		Name:               hs.name,
		GitHash:            hs.gitHash,
		ContainerStartedAt: hs.containerStartedAt,
		Dependencies: HealthDependencies{
			Database: hs.databaseOK,
		},
		CachedAt: hs.cachedAt,
	}
}
