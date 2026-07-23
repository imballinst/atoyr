package leaderboard

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type Entry struct {
	UserID        string
	Score         int
	TotalAttempts int
	Accuracy      float64
	FinishedAt    time.Time
}

type ActiveSession struct {
	ID        string
	UserID    string
	Mode      string
	StartedAt time.Time
}

type Store interface {
	AddSession(id, userID, mode string, startedAt time.Time) error
	RemoveSession(id string) error
	ListActiveSessions() []ActiveSession
	UpsertEntry(mode string, entry Entry) error
	ListEntries(mode string, limit, offset int) ([]Entry, int, error)
	Percentile(mode, userID string) (float64, error)
}

type InMemoryStore struct {
	mu       sync.RWMutex
	entries  map[string]map[string]Entry
	sessions map[string]ActiveSession
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		entries:  make(map[string]map[string]Entry),
		sessions: make(map[string]ActiveSession),
	}
}

func (s *InMemoryStore) AddSession(id, userID, mode string, startedAt time.Time) error {
	if id == "" {
		return errors.New("session id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[id] = ActiveSession{
		ID:        id,
		UserID:    userID,
		Mode:      mode,
		StartedAt: startedAt,
	}
	return nil
}

func (s *InMemoryStore) RemoveSession(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
	return nil
}

func (s *InMemoryStore) ListActiveSessions() []ActiveSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]ActiveSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

func (s *InMemoryStore) UpsertEntry(mode string, entry Entry) error {
	if mode == "" {
		return errors.New("mode is required")
	}
	if entry.UserID == "" {
		return errors.New("user id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.entries[mode] == nil {
		s.entries[mode] = make(map[string]Entry)
	}
	s.entries[mode][entry.UserID] = entry
	return nil
}

func (s *InMemoryStore) ListEntries(mode string, limit, offset int) ([]Entry, int, error) {
	if mode == "" {
		return nil, 0, errors.New("mode is required")
	}
	if limit < 0 {
		return nil, 0, errors.New("limit cannot be negative")
	}
	if offset < 0 {
		return nil, 0, errors.New("offset cannot be negative")
	}

	all := s.sortedEntries(mode)
	total := len(all)

	if offset > total {
		return []Entry{}, total, nil
	}

	end := offset + limit
	if end > total || limit == 0 {
		end = total
	}

	return all[offset:end], total, nil
}

func (s *InMemoryStore) Percentile(mode, userID string) (float64, error) {
	if mode == "" {
		return 0, errors.New("mode is required")
	}
	if userID == "" {
		return 0, errors.New("user id is required")
	}

	all := s.sortedEntries(mode)

	var current Entry
	found := false
	for _, entry := range all {
		if entry.UserID == userID {
			current = entry
			found = true
			break
		}
	}
	if !found {
		return 0, ErrUserNotFound
	}

	totalEligible := len(all) - 1
	if totalEligible <= 0 {
		return 0, nil
	}

	worse := 0
	for _, entry := range all {
		if entry.UserID == userID {
			continue
		}
		if ranksBelow(entry, current) {
			worse++
		}
	}

	return float64(worse) / float64(totalEligible) * 100, nil
}

func (s *InMemoryStore) sortedEntries(mode string) []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	modeEntries := s.entries[mode]
	all := make([]Entry, 0, len(modeEntries))
	for _, entry := range modeEntries {
		all = append(all, entry)
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].Score != all[j].Score {
			return all[i].Score > all[j].Score
		}
		if all[i].Accuracy != all[j].Accuracy {
			return all[i].Accuracy > all[j].Accuracy
		}
		return all[i].FinishedAt.Before(all[j].FinishedAt)
	})

	return all
}

func ranksBelow(candidate, reference Entry) bool {
	if candidate.Score != reference.Score {
		return candidate.Score < reference.Score
	}
	if candidate.Accuracy != reference.Accuracy {
		return candidate.Accuracy < reference.Accuracy
	}
	return candidate.FinishedAt.After(reference.FinishedAt)
}
