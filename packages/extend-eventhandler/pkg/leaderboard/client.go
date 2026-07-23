package leaderboard

import (
	"context"
	"time"

	leaderboardv1 "extend-event-handler/pkg/pb/leaderboard/v1"

	"google.golang.org/grpc"
)

type Client interface {
	AddSession(ctx context.Context, sessionID, userID, mode string, startedAt time.Time) error
	RemoveSession(ctx context.Context, sessionID string) error
	UpsertEntry(ctx context.Context, mode, userID string, score int, totalAttempts int, accuracy float64, finishedAt time.Time) error
}

type grpcClient struct {
	client leaderboardv1.LeaderboardServiceClient
}

func NewClient(conn grpc.ClientConnInterface) Client {
	return &grpcClient{
		client: leaderboardv1.NewLeaderboardServiceClient(conn),
	}
}

func (c *grpcClient) AddSession(ctx context.Context, sessionID, userID, mode string, startedAt time.Time) error {
	_, err := c.client.AddSession(ctx, &leaderboardv1.AddSessionRequest{
		SessionId: sessionID,
		UserId:    userID,
		Mode:      mode,
		StartedAt: startedAt.UnixMilli(),
	})
	return err
}

func (c *grpcClient) RemoveSession(ctx context.Context, sessionID string) error {
	_, err := c.client.RemoveSession(ctx, &leaderboardv1.RemoveSessionRequest{
		SessionId: sessionID,
	})
	return err
}

func (c *grpcClient) UpsertEntry(ctx context.Context, mode, userID string, score int, totalAttempts int, accuracy float64, finishedAt time.Time) error {
	_, err := c.client.UpsertEntry(ctx, &leaderboardv1.UpsertEntryRequest{
		Mode:          mode,
		UserId:        userID,
		Score:         int32(score),
		TotalAttempts: int32(totalAttempts),
		Accuracy:      accuracy,
		FinishedAt:    finishedAt.UnixMilli(),
	})
	return err
}
