package services

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"atoyr/server/internal/platform/accelbyte"

	"github.com/AccelByte/accelbyte-go-sdk/cloudsave-sdk/pkg/cloudsaveclient/admin_player_record"
	"github.com/AccelByte/accelbyte-go-sdk/cloudsave-sdk/pkg/cloudsaveclientmodels"
	"github.com/AccelByte/accelbyte-go-sdk/iam-sdk/pkg/iamclient/o_auth2_0"
	"github.com/AccelByte/accelbyte-go-sdk/iam-sdk/pkg/iamclient/o_auth2_0_extension"
	"github.com/AccelByte/accelbyte-go-sdk/leaderboard-sdk/pkg/leaderboardclient/leaderboard_data_v3"
	"github.com/AccelByte/accelbyte-go-sdk/leaderboard-sdk/pkg/leaderboardclientmodels"
	"github.com/AccelByte/accelbyte-go-sdk/social-sdk/pkg/socialclient/user_statistic"
	"github.com/AccelByte/accelbyte-go-sdk/social-sdk/pkg/socialclientmodels"
)

const (
	cloudSaveRoundKeyPrefix = "atoyr:round"
)

// CloudSaveRoundState mirrors the round state stored in AGS Cloud Save.
type CloudSaveRoundState struct {
	Mode                     string   `json:"mode"`
	Phase                    string   `json:"phase"`
	AutoVoice                bool     `json:"autoVoice"`
	ItemsUsed                []string `json:"itemsUsed"`
	UsedWords                []string `json:"usedWords"`
	CurrentWord              string   `json:"currentWord"`
	CurrentScrambledWord     string   `json:"currentScrambledWord"`
	CurrentWordDefinition    string   `json:"currentWordDefinition"`
	CurrentWordToken         string   `json:"currentWordToken"`
	Score                    int32    `json:"score"`
	TotalAttempts            int32    `json:"totalAttempts"`
	Accuracy                 float32  `json:"accuracy"`
	CorrectAttemptTimestamps []string `json:"correctAttemptTimestamps"`
	EndsAt                   int64    `json:"endsAt"`
	CreatedAt                int64    `json:"createdAt"`
	UpdatedAt                int64    `json:"updatedAt"`
}

type AGSSyncService struct {
	client *accelbyte.Client
}

func NewAGSSyncService(client *accelbyte.Client) *AGSSyncService {
	return &AGSSyncService{client: client}
}

func (a *AGSSyncService) Enabled() bool {
	return a != nil && a.client != nil
}

func (a *AGSSyncService) CreateHeadlessAccount() (string, string, error) {
	input := &o_auth2_0_extension.GenerateTokenByNewHeadlessAccountV3Params{}
	resp, err := a.client.OAuth20ExtensionService.GenerateTokenByNewHeadlessAccountV3Short(input)
	if err != nil {
		return "", "", fmt.Errorf("failed to create headless account: %w", err)
	}
	if resp == nil || resp.AccessToken == nil || resp.UserID == "" {
		return "", "", fmt.Errorf("empty headless account response")
	}
	return resp.UserID, *resp.AccessToken, nil
}

func (a *AGSSyncService) ValidatePlayerToken(token string) (string, error) {
	input := &o_auth2_0.TokenIntrospectionV3Params{Token: token}
	resp, err := a.client.OAuth20Service.TokenIntrospectionV3Short(input)
	if err != nil {
		return "", fmt.Errorf("token introspection failed: %w", err)
	}
	if resp == nil || resp.Active == nil || !*resp.Active {
		return "", fmt.Errorf("token is not active")
	}
	return resp.Sub, nil
}

func (a *AGSSyncService) SaveRoundState(userID, roundID string, state CloudSaveRoundState) error {
	state.UpdatedAt = time.Now().UnixMilli()
	body := cloudsaveclientmodels.ModelsPlayerRecordRequest(state)
	input := &admin_player_record.AdminPutPlayerRecordHandlerV1Params{
		Namespace: a.client.Config.Namespace,
		UserID:    userID,
		Key:       roundRecordKey(roundID),
		Body:      body,
	}
	_, err := a.client.CloudSaveService.AdminPutPlayerRecordHandlerV1Short(input)
	if err != nil {
		return fmt.Errorf("failed to save round state to Cloud Save: %w", err)
	}
	return nil
}

func (a *AGSSyncService) LoadRoundState(userID, roundID string) (*CloudSaveRoundState, error) {
	input := &admin_player_record.AdminGetPlayerRecordHandlerV1Params{
		Namespace: a.client.Config.Namespace,
		UserID:    userID,
		Key:       roundRecordKey(roundID),
	}
	resp, err := a.client.CloudSaveService.AdminGetPlayerRecordHandlerV1Short(input)
	if err != nil {
		return nil, fmt.Errorf("failed to load round state from Cloud Save: %w", err)
	}
	if resp == nil || resp.Value == nil {
		return nil, fmt.Errorf("round state not found in Cloud Save")
	}
	value, ok := resp.Value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected round state type")
	}
	return mapToRoundState(value)
}

func (a *AGSSyncService) PostRoundStats(userID, mode string, score, attempts int32, accuracy float32, durationSeconds int32) error {
	composite := calculateCompositeScore(score, accuracy, durationSeconds)
	updates := []*socialclientmodels.BulkUserStatItemUpdate{
		statUpdate(userID, fmt.Sprintf("atoyr-score-%s", mode), float64(score)),
		statUpdate(userID, fmt.Sprintf("atoyr-attempts-%s", mode), float64(attempts)),
		statUpdate(userID, fmt.Sprintf("atoyr-accuracy-%s", mode), float64(accuracy)),
		statUpdate(userID, fmt.Sprintf("atoyr-composite-%s", mode), float64(composite)),
	}
	input := &user_statistic.BulkUpdateUserStatItemV2Params{
		Namespace: a.client.Config.Namespace,
		Body:      updates,
	}
	_, err := a.client.UserStatisticService.BulkUpdateUserStatItemV2Short(input)
	if err != nil {
		return fmt.Errorf("failed to post round stats: %w", err)
	}
	return nil
}

func (a *AGSSyncService) GetLeaderboard(mode string, limit, offset int) ([]LeaderboardEntry, int64, error) {
	leaderboardCode := fmt.Sprintf("atoyr-leaderboard-%s", mode)
	limit64 := int64(limit)
	offset64 := int64(offset)
	input := &leaderboard_data_v3.GetAllTimeLeaderboardRankingAdminV3Params{
		Namespace:       a.client.Config.Namespace,
		LeaderboardCode: leaderboardCode,
		Limit:           &limit64,
		Offset:          &offset64,
	}
	resp, err := a.client.LeaderboardDataService.GetAllTimeLeaderboardRankingAdminV3Short(input)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch AGS leaderboard: %w", err)
	}
	return a.transformLeaderboardResponse(resp, offset)
}

func (a *AGSSyncService) GetUserRank(userID, mode string) (int64, int64, error) {
	leaderboardCode := fmt.Sprintf("atoyr-leaderboard-%s", mode)
	input := &leaderboard_data_v3.GetUserRankingAdminV3Params{
		Namespace:       a.client.Config.Namespace,
		LeaderboardCode: leaderboardCode,
		UserID:          userID,
	}
	resp, err := a.client.LeaderboardDataService.GetUserRankingAdminV3Short(input)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch user rank: %w", err)
	}
	if resp == nil || resp.AllTime == nil || resp.AllTime.Rank == nil {
		return 0, 0, fmt.Errorf("user rank not found")
	}
	total, err := a.countLeaderboardEntries(leaderboardCode)
	if err != nil {
		return 0, 0, err
	}
	return *resp.AllTime.Rank, total, nil
}

func (a *AGSSyncService) transformLeaderboardResponse(resp *leaderboardclientmodels.ModelsGetLeaderboardRankingResp, offset int) ([]LeaderboardEntry, int64, error) {
	if resp == nil {
		return nil, 0, nil
	}
	entries := make([]LeaderboardEntry, 0, len(resp.Data))
	for i, item := range resp.Data {
		var score int32
		if item.Point != nil {
			score = int32(*item.Point)
		}
		entry := LeaderboardEntry{
			ID:    safeString(item.UserID),
			Rank:  int32(i + 1 + offset),
			Score: score,
		}
		entries = append(entries, entry)
	}
	total := int64(len(resp.Data))
	return entries, total, nil
}

func (a *AGSSyncService) countLeaderboardEntries(leaderboardCode string) (int64, error) {
	limit := int64(10000)
	input := &leaderboard_data_v3.GetAllTimeLeaderboardRankingAdminV3Params{
		Namespace:       a.client.Config.Namespace,
		LeaderboardCode: leaderboardCode,
		Limit:           &limit,
	}
	resp, err := a.client.LeaderboardDataService.GetAllTimeLeaderboardRankingAdminV3Short(input)
	if err != nil {
		return 0, fmt.Errorf("failed to count leaderboard entries: %w", err)
	}
	if resp == nil {
		return 0, nil
	}
	return int64(len(resp.Data)), nil
}

func roundRecordKey(roundID string) string {
	return fmt.Sprintf("%s:%s", cloudSaveRoundKeyPrefix, roundID)
}

func statUpdate(userID, statCode string, value float64) *socialclientmodels.BulkUserStatItemUpdate {
	strategy := "OVERRIDE"
	return &socialclientmodels.BulkUserStatItemUpdate{
		StatCode:       &statCode,
		UpdateStrategy: &strategy,
		UserID:         &userID,
		Value:          &value,
	}
}

func calculateCompositeScore(score int32, accuracy float32, durationSeconds int32) int64 {
	if durationSeconds < 0 {
		durationSeconds = 0
	}
	return int64(score)*1000000 + int64(math.Round(float64(accuracy)*10000)) - int64(durationSeconds)
}

func mapToRoundState(value map[string]interface{}) (*CloudSaveRoundState, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal round state: %w", err)
	}
	var state CloudSaveRoundState
	if err := json.Unmarshal(bytes, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal round state: %w", err)
	}
	return &state, nil
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Async helpers

func (a *AGSSyncService) SaveRoundStateAsync(userID, roundID string, state CloudSaveRoundState) {
	go func() {
		if err := a.SaveRoundState(userID, roundID, state); err != nil {
			log.Printf("async Cloud Save write failed for round %s: %v", roundID, err)
		}
	}()
}

func (a *AGSSyncService) PostRoundStatsAsync(userID, mode string, score, attempts int32, accuracy float32, durationSeconds int32) {
	go func() {
		if err := a.PostRoundStats(userID, mode, score, attempts, accuracy, durationSeconds); err != nil {
			log.Printf("async AGS stats post failed for user %s: %v", userID, err)
		}
	}()
}
