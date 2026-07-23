package leaderboard

import (
	"context"
	"math"
	"time"

	leaderboardpb "extend-leaderboard/pkg/pb/leaderboard/v1"
	"extend-leaderboard/pkg/storage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ leaderboardpb.LeaderboardServiceServer = (*LeaderboardServiceServer)(nil)

type LeaderboardServiceServer struct {
	leaderboardpb.UnimplementedLeaderboardServiceServer
	store storage.LeaderboardStore
}

func NewLeaderboardServiceServer(store storage.LeaderboardStore) *LeaderboardServiceServer {
	return &LeaderboardServiceServer{store: store}
}

func (s *LeaderboardServiceServer) GetLeaderboard(ctx context.Context, req *leaderboardpb.GetLeaderboardRequest) (*leaderboardpb.GetLeaderboardResponse, error) {
	mode := req.GetMode()
	if mode == "" {
		return nil, status.Error(codes.InvalidArgument, "mode is required")
	}

	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 10
	}
	offset := int(req.GetOffset())
	if offset < 0 {
		offset = 0
	}

	entries, total, err := s.store.ListEntries(ctx, mode, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbEntries := make([]*leaderboardpb.LeaderboardEntry, len(entries))
	for i, entry := range entries {
		pbEntries[i] = &leaderboardpb.LeaderboardEntry{
			Rank:          int32(i + 1 + offset),
			UserId:        entry.UserID,
			Score:         int32(entry.Score),
			TotalAttempts: int32(entry.TotalAttempts),
			Accuracy:      entry.Accuracy,
			Timestamp:     entry.FinishedAt.UnixMilli(),
		}
	}

	return &leaderboardpb.GetLeaderboardResponse{
		Entries: pbEntries,
		Total:   int32(total),
	}, nil
}

func (s *LeaderboardServiceServer) GetPercentile(ctx context.Context, req *leaderboardpb.GetPercentileRequest) (*leaderboardpb.GetPercentileResponse, error) {
	mode := req.GetMode()
	if mode == "" {
		return nil, status.Error(codes.InvalidArgument, "mode is required")
	}

	userID := req.GetUserId()
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "userId is required")
	}

	percentile, err := s.store.Percentile(ctx, mode, userID)
	if err != nil {
		if err == storage.ErrUserNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if math.IsNaN(percentile) {
		percentile = 0
	}

	return &leaderboardpb.GetPercentileResponse{Percentile: percentile}, nil
}

func (s *LeaderboardServiceServer) AddSession(ctx context.Context, req *leaderboardpb.AddSessionRequest) (*leaderboardpb.AddSessionResponse, error) {
	if req.GetSessionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	err := s.store.AddSession(ctx, req.GetSessionId(), req.GetUserId(), req.GetMode(), time.UnixMilli(req.GetStartedAt()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &leaderboardpb.AddSessionResponse{}, nil
}

func (s *LeaderboardServiceServer) RemoveSession(ctx context.Context, req *leaderboardpb.RemoveSessionRequest) (*leaderboardpb.RemoveSessionResponse, error) {
	if req.GetSessionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	err := s.store.RemoveSession(ctx, req.GetSessionId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &leaderboardpb.RemoveSessionResponse{}, nil
}

func (s *LeaderboardServiceServer) UpsertEntry(ctx context.Context, req *leaderboardpb.UpsertEntryRequest) (*leaderboardpb.UpsertEntryResponse, error) {
	if req.GetMode() == "" {
		return nil, status.Error(codes.InvalidArgument, "mode is required")
	}

	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.store.UpsertEntry(ctx, req.GetMode(), storage.Entry{
		UserID:        req.GetUserId(),
		Score:         int(req.GetScore()),
		TotalAttempts: int(req.GetTotalAttempts()),
		Accuracy:      req.GetAccuracy(),
		FinishedAt:    time.UnixMilli(req.GetFinishedAt()),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &leaderboardpb.UpsertEntryResponse{}, nil
}
