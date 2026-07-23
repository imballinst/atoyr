package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	Close(ctx context.Context) error
}

// LeaderboardStore defines the query methods needed by the leaderboard service.
type LeaderboardStore interface {
	AddSession(ctx context.Context, sessionID, userID, mode string, startedAt time.Time) error
	RemoveSession(ctx context.Context, sessionID string) error
	ListActiveSessions(ctx context.Context) ([]ActiveSession, error)
	UpsertEntry(ctx context.Context, mode string, entry Entry) error
	ListEntries(ctx context.Context, mode string, limit, offset int) ([]Entry, int, error)
	Percentile(ctx context.Context, mode, userID string) (float64, error)
}

type PostgreSQLStorage struct {
	pool *pgxpool.Pool
}

func NewPostgreSQLStorage(connectionString string) (*PostgreSQLStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PostgreSQL connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL connection pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	storage := &PostgreSQLStorage{pool: pool}

	if err = storage.initializeSchema(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return storage, nil
}

func (p *PostgreSQLStorage) initializeSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS leaderboard_entries (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id TEXT NOT NULL,
		mode TEXT NOT NULL,
		score INTEGER NOT NULL,
		total_attempts INTEGER NOT NULL DEFAULT 0,
		accuracy REAL NOT NULL DEFAULT 0,
		finished_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE
	);

	CREATE INDEX IF NOT EXISTS idx_leaderboard_entries_mode_score
		ON leaderboard_entries (mode, score DESC);

	CREATE INDEX IF NOT EXISTS idx_leaderboard_entries_user_mode
		ON leaderboard_entries (user_id, mode);

	CREATE UNIQUE INDEX IF NOT EXISTS idx_leaderboard_entries_user_mode_unique
		ON leaderboard_entries (user_id, mode);

	CREATE TABLE IF NOT EXISTS active_sessions (
		session_id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		mode TEXT NOT NULL,
		started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);
	`

	_, err := p.pool.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

func (p *PostgreSQLStorage) Close(ctx context.Context) error {
	p.pool.Close()
	return nil
}

func (p *PostgreSQLStorage) Pool() *pgxpool.Pool {
	return p.pool
}
