package accelbyte

import (
	"fmt"
	"os"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/cloudsave"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/leaderboard"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/social"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils/auth"
)

type Config struct {
	BaseURL      string
	Namespace    string
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

func NewConfigFromEnv() (*Config, error) {
	cfg := &Config{
		BaseURL:      os.Getenv("AGS_BASE_URL"),
		Namespace:    os.Getenv("AGS_NAMESPACE"),
		ClientID:     os.Getenv("AGS_CLIENT_ID"),
		ClientSecret: os.Getenv("AGS_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("AGS_ADMIN_REDIRECT_URI"),
	}
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("AGS_BASE_URL is required")
	}
	if cfg.Namespace == "" {
		return nil, fmt.Errorf("AGS_NAMESPACE is required")
	}
	if cfg.ClientID == "" {
		return nil, fmt.Errorf("AGS_CLIENT_ID is required")
	}
	if cfg.ClientSecret == "" {
		return nil, fmt.Errorf("AGS_CLIENT_SECRET is required")
	}
	return cfg, nil
}

type configRepository struct {
	cfg *Config
}

func (r *configRepository) GetClientId() string {
	return r.cfg.ClientID
}

func (r *configRepository) GetClientSecret() string {
	return r.cfg.ClientSecret
}

func (r *configRepository) GetJusticeBaseUrl() string {
	return r.cfg.BaseURL
}

type Client struct {
	Config                  *Config
	ConfigRepository        repository.ConfigRepository
	TokenRepository         repository.TokenRepository
	OAuth20Service          *iam.OAuth20Service
	OAuth20ExtensionService *iam.OAuth20ExtensionService
	CloudSaveService        *cloudsave.AdminPlayerRecordService
	UserStatisticService    *social.UserStatisticService
	LeaderboardService      *leaderboard.UserDataV3Service
	LeaderboardDataService  *leaderboard.LeaderboardDataV3Service
}

func NewClient(cfg *Config) (*Client, error) {
	configRepo := &configRepository{cfg: cfg}
	tokenRepo := auth.DefaultTokenRepositoryImpl()

	oauth := &iam.OAuth20Service{
		Client:           factory.NewIamClient(configRepo),
		ConfigRepository: configRepo,
		TokenRepository:  tokenRepo,
	}

	oauthExt := &iam.OAuth20ExtensionService{
		Client:           factory.NewIamClient(configRepo),
		ConfigRepository: configRepo,
		TokenRepository:  tokenRepo,
	}

	cloudSaveSvc := &cloudsave.AdminPlayerRecordService{
		Client:           factory.NewCloudsaveClient(configRepo),
		ConfigRepository: configRepo,
		TokenRepository:  tokenRepo,
	}

	userStatSvc := &social.UserStatisticService{
		Client:           factory.NewSocialClient(configRepo),
		ConfigRepository: configRepo,
		TokenRepository:  tokenRepo,
	}

	leaderboardSvc := &leaderboard.UserDataV3Service{
		Client:           factory.NewLeaderboardClient(configRepo),
		ConfigRepository: configRepo,
		TokenRepository:  tokenRepo,
	}

	leaderboardDataSvc := &leaderboard.LeaderboardDataV3Service{
		Client:           factory.NewLeaderboardClient(configRepo),
		ConfigRepository: configRepo,
		TokenRepository:  tokenRepo,
	}

	return &Client{
		Config:                  cfg,
		ConfigRepository:        configRepo,
		TokenRepository:         tokenRepo,
		OAuth20Service:          oauth,
		OAuth20ExtensionService: oauthExt,
		CloudSaveService:        cloudSaveSvc,
		UserStatisticService:    userStatSvc,
		LeaderboardService:      leaderboardSvc,
		LeaderboardDataService:  leaderboardDataSvc,
	}, nil
}

func (c *Client) LoginClientCredentials() error {
	return c.OAuth20Service.LoginClient(nil, nil)
}
