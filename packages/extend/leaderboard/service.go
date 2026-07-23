package leaderboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type LeaderboardResponse struct {
	Entries []LeaderboardEntryResponse `json:"entries"`
	Total   int                        `json:"total"`
}

type LeaderboardEntryResponse struct {
	Rank          int     `json:"rank"`
	UserID        string  `json:"userId"`
	Score         int     `json:"score"`
	TotalAttempts int     `json:"totalAttempts"`
	Accuracy      float64 `json:"accuracy"`
	Timestamp     int64   `json:"timestamp"`
}

type PercentileResponse struct {
	Percentile float64 `json:"percentile"`
}

type ActiveSessionsResponse struct {
	Sessions []ActiveSessionResponse `json:"sessions"`
}

type ActiveSessionResponse struct {
	SessionID string `json:"sessionId"`
	UserID    string `json:"userId"`
	Mode      string `json:"mode"`
	StartedAt int64  `json:"startedAt"`
}

type Server struct {
	store Store
}

func NewServer(store Store) *Server {
	return &Server{store: store}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /leaderboard", s.getLeaderboard)
	mux.HandleFunc("GET /leaderboard/percentile", s.getPercentile)
	mux.HandleFunc("GET /active-sessions", s.getActiveSessions)
	return mux
}

func (s *Server) getLeaderboard(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")
	if mode == "" {
		http.Error(w, "mode is required", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	entries, total, err := s.store.ListEntries(mode, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := LeaderboardResponse{
		Entries: make([]LeaderboardEntryResponse, len(entries)),
		Total:   total,
	}
	for i, entry := range entries {
		response.Entries[i] = LeaderboardEntryResponse{
			Rank:          i + 1 + offset,
			UserID:        entry.UserID,
			Score:         entry.Score,
			TotalAttempts: entry.TotalAttempts,
			Accuracy:      entry.Accuracy,
			Timestamp:     entry.FinishedAt.UnixMilli(),
		}
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) getPercentile(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")
	if mode == "" {
		http.Error(w, "mode is required", http.StatusBadRequest)
		return
	}

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	percentile, err := s.store.Percentile(mode, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, PercentileResponse{Percentile: percentile})
}

func (s *Server) getActiveSessions(w http.ResponseWriter, r *http.Request) {
	sessions := s.store.ListActiveSessions()
	response := ActiveSessionsResponse{
		Sessions: make([]ActiveSessionResponse, len(sessions)),
	}
	for i, session := range sessions {
		response.Sessions[i] = ActiveSessionResponse{
			SessionID: session.ID,
			UserID:    session.UserID,
			Mode:      session.Mode,
			StartedAt: session.StartedAt.UnixMilli(),
		}
	}

	writeJSON(w, http.StatusOK, response)
}

var ErrUserNotFound = errors.New("user not found")

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
