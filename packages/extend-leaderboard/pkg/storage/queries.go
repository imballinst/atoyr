package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type ActiveSession struct {
	ID        string
	UserID    string
	Mode      string
	StartedAt time.Time
}

type Entry struct {
	UserID        string
	Score         int
	TotalAttempts int
	Accuracy      float64
	FinishedAt    time.Time
}

var ErrUserNotFound = errors.New("user not found")

func (p *PostgreSQLStorage) AddSession(ctx context.Context, sessionID, userID, mode string, startedAt time.Time) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO active_sessions (session_id, user_id, mode, started_at) VALUES ($1, $2, $3, $4)`,
		sessionID, userID, mode, startedAt,
	)
	return err
}

func (p *PostgreSQLStorage) RemoveSession(ctx context.Context, sessionID string) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM active_sessions WHERE session_id = $1`, sessionID)
	return err
}

func (p *PostgreSQLStorage) ListActiveSessions(ctx context.Context) ([]ActiveSession, error) {
	rows, err := p.pool.Query(ctx, `SELECT session_id, user_id, mode, started_at FROM active_sessions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []ActiveSession
	for rows.Next() {
		var s ActiveSession
		if err := rows.Scan(&s.ID, &s.UserID, &s.Mode, &s.StartedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (p *PostgreSQLStorage) UpsertEntry(ctx context.Context, mode string, entry Entry) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO leaderboard_entries (user_id, mode, score, total_attempts, accuracy, finished_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (user_id, mode) DO UPDATE SET
		   score = EXCLUDED.score,
		   total_attempts = EXCLUDED.total_attempts,
		   accuracy = EXCLUDED.accuracy,
		   finished_at = EXCLUDED.finished_at,
		   updated_at = NOW()`,
		entry.UserID, mode, entry.Score, entry.TotalAttempts, entry.Accuracy, entry.FinishedAt,
	)
	return err
}

func (p *PostgreSQLStorage) ListEntries(ctx context.Context, mode string, limit, offset int) ([]Entry, int, error) {
	var total int
	err := p.pool.QueryRow(ctx, `SELECT COUNT(*) FROM leaderboard_entries WHERE mode = $1`, mode).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := p.pool.Query(ctx,
		`SELECT user_id, score, total_attempts, accuracy, finished_at
		 FROM leaderboard_entries
		 WHERE mode = $1
		 ORDER BY score DESC, accuracy DESC, finished_at ASC
		 LIMIT $2 OFFSET $3`,
		mode, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.UserID, &e.Score, &e.TotalAttempts, &e.Accuracy, &e.FinishedAt); err != nil {
			return nil, 0, err
		}
		entries = append(entries, e)
	}
	return entries, total, rows.Err()
}

func (p *PostgreSQLStorage) Percentile(ctx context.Context, mode, userID string) (float64, error) {
	var userEntry Entry
	err := p.pool.QueryRow(ctx,
		`SELECT user_id, score, total_attempts, accuracy, finished_at
		 FROM leaderboard_entries
		 WHERE mode = $1 AND user_id = $2`,
		mode, userID,
	).Scan(&userEntry.UserID, &userEntry.Score, &userEntry.TotalAttempts, &userEntry.Accuracy, &userEntry.FinishedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrUserNotFound
		}
		return 0, err
	}

	var totalEligible int
	err = p.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM leaderboard_entries WHERE mode = $1 AND user_id != $2`,
		mode, userID,
	).Scan(&totalEligible)
	if err != nil {
		return 0, err
	}

	if totalEligible == 0 {
		return 0, nil
	}

	var worse int
	err = p.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM leaderboard_entries
		 WHERE mode = $1 AND user_id != $2
		 AND (
		   score < $3
		   OR (score = $3 AND accuracy < $4)
		   OR (score = $3 AND accuracy = $4 AND finished_at > $5)
		 )`,
		mode, userID, userEntry.Score, userEntry.Accuracy, userEntry.FinishedAt,
	).Scan(&worse)
	if err != nil {
		return 0, err
	}

	return float64(worse) / float64(totalEligible) * 100, nil
}
