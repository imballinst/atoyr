package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const (
	defaultExtendLeaderboardBase = "http://localhost:8000/leaderboard"
)

type LeaderboardReadClient interface {
	GetLeaderboard(mode string, limit, offset int) ([]LeaderboardEntry, int64, error)
	GetPercentile(mode, userID string) (float64, error)
	GetTotalEntries(mode string) (int64, error)
	UpsertEntry(mode, userID string, score, attempts int32, accuracy float32, finishedAt int64) error
}

type ExtendLeaderboardClient struct {
	baseURL string
	client  *http.Client
}

func NewExtendLeaderboardClient(baseURL string) *ExtendLeaderboardClient {
	if baseURL == "" {
		baseURL = defaultExtendLeaderboardBase
	}
	return &ExtendLeaderboardClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

type extendLeaderboardResponse struct {
	Entries []extendLeaderboardEntry `json:"entries"`
	Total   int                      `json:"total"`
}

type extendLeaderboardEntry struct {
	Rank          int32   `json:"rank"`
	UserID        string  `json:"userId"`
	Score         int32   `json:"score"`
	TotalAttempts int32   `json:"totalAttempts"`
	Accuracy      float64 `json:"accuracy"`
	Timestamp     int64   `json:"timestamp"`
}

type extendPercentileResponse struct {
	Percentile float64 `json:"percentile"`
}

type upsertEntryBody struct {
	Mode          string  `json:"mode"`
	UserID        string  `json:"userId"`
	Score         int32   `json:"score"`
	TotalAttempts int32   `json:"totalAttempts"`
	Accuracy      float64 `json:"accuracy"`
	FinishedAt    int64   `json:"finishedAt"`
}

func (c *ExtendLeaderboardClient) GetLeaderboard(mode string, limit, offset int) ([]LeaderboardEntry, int64, error) {
	u, err := url.Parse(c.baseURL + "/v1/leaderboard")
	if err != nil {
		return nil, 0, fmt.Errorf("invalid base URL: %w", err)
	}
	q := u.Query()
	q.Set("mode", mode)
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	u.RawQuery = q.Encode()

	resp, err := c.client.Get(u.String())
	if err != nil {
		log.Printf("extend-leaderboard request failed: %s: %v", u.String(), err)
		return nil, 0, fmt.Errorf("failed to call extend-leaderboard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("extend-leaderboard returned status %d for %s: %s", resp.StatusCode, u.String(), string(bodyBytes))
		return nil, 0, fmt.Errorf("extend-leaderboard returned status %d", resp.StatusCode)
	}

	var body extendLeaderboardResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		log.Printf("extend-leaderboard decode failed for %s: %v", u.String(), err)
		return nil, 0, fmt.Errorf("failed to decode extend-leaderboard response: %w", err)
	}

	entries := make([]LeaderboardEntry, len(body.Entries))
	for i, e := range body.Entries {
		entries[i] = LeaderboardEntry{
			ID:            e.UserID,
			Rank:          e.Rank,
			Score:         e.Score,
			TotalAttempts: e.TotalAttempts,
			Accuracy:      float32(e.Accuracy),
			Timestamp:     e.Timestamp,
		}
	}

	return entries, int64(body.Total), nil
}

func (c *ExtendLeaderboardClient) GetPercentile(mode, userID string) (float64, error) {
	u, err := url.Parse(c.baseURL + "/v1/leaderboard/percentile")
	if err != nil {
		return 0, fmt.Errorf("invalid base URL: %w", err)
	}
	q := u.Query()
	q.Set("mode", mode)
	q.Set("userId", userID)
	u.RawQuery = q.Encode()

	resp, err := c.client.Get(u.String())
	if err != nil {
		log.Printf("extend-leaderboard request failed: %s: %v", u.String(), err)
		return 0, fmt.Errorf("failed to call extend-leaderboard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return 0, nil
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("extend-leaderboard returned status %d for %s: %s", resp.StatusCode, u.String(), string(bodyBytes))
		return 0, fmt.Errorf("extend-leaderboard returned status %d", resp.StatusCode)
	}

	var body extendPercentileResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		log.Printf("extend-leaderboard decode failed for %s: %v", u.String(), err)
		return 0, fmt.Errorf("failed to decode extend-leaderboard response: %w", err)
	}

	return body.Percentile, nil
}

func (c *ExtendLeaderboardClient) UpsertEntry(mode, userID string, score, attempts int32, accuracy float32, finishedAt int64) error {
	u, err := url.Parse(c.baseURL + "/v1/leaderboard/entry")
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	body := upsertEntryBody{
		Mode:          mode,
		UserID:        userID,
		Score:         score,
		TotalAttempts: attempts,
		Accuracy:      float64(accuracy),
		FinishedAt:    finishedAt,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(u.String(), "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("extend-leaderboard upsert entry request failed: %s: %v", u.String(), err)
		return fmt.Errorf("failed to call extend-leaderboard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("extend-leaderboard upsert entry returned status %d for %s: %s", resp.StatusCode, u.String(), string(respBody))
		return fmt.Errorf("extend-leaderboard returned status %d", resp.StatusCode)
	}

	return nil
}

func (c *ExtendLeaderboardClient) GetTotalEntries(mode string) (int64, error) {
	u, err := url.Parse(c.baseURL + "/v1/leaderboard")
	if err != nil {
		return 0, fmt.Errorf("invalid base URL: %w", err)
	}
	q := u.Query()
	q.Set("mode", mode)
	q.Set("limit", "1")
	q.Set("offset", "0")
	u.RawQuery = q.Encode()

	resp, err := c.client.Get(u.String())
	if err != nil {
		log.Printf("extend-leaderboard request failed: %s: %v", u.String(), err)
		return 0, fmt.Errorf("failed to call extend-leaderboard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("extend-leaderboard returned status %d for %s: %s", resp.StatusCode, u.String(), string(bodyBytes))
		return 0, fmt.Errorf("extend-leaderboard returned status %d", resp.StatusCode)
	}

	var body extendLeaderboardResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		log.Printf("extend-leaderboard decode failed for %s: %v", u.String(), err)
		return 0, fmt.Errorf("failed to decode extend-leaderboard response: %w", err)
	}

	return int64(body.Total), nil
}
