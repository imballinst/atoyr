package service

import (
	pb "extend-leaderboard/pkg/pb"
	"extend-leaderboard/pkg/storage"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository"
)

var _ pb.ServiceServer = (*MyServiceServer)(nil)

type MyServiceServer struct {
	pb.UnimplementedServiceServer
	tokenRepo   repository.TokenRepository
	configRepo  repository.ConfigRepository
	refreshRepo repository.RefreshTokenRepository
	storage     storage.Storage
}

func NewMyServiceServer(
	tokenRepo repository.TokenRepository,
	configRepo repository.ConfigRepository,
	refreshRepo repository.RefreshTokenRepository,
	storage storage.Storage,
) *MyServiceServer {
	return &MyServiceServer{
		tokenRepo:   tokenRepo,
		configRepo:  configRepo,
		refreshRepo: refreshRepo,
		storage:     storage,
	}
}
