-- 001_initial_schema.sql
-- Initial schema for Atoyr server

CREATE TABLE IF NOT EXISTS session_entities (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    ends_at DATETIME NOT NULL,
    phase TEXT NOT NULL DEFAULT 'idle',
    score INTEGER NOT NULL DEFAULT 0,
    total_attempts INTEGER NOT NULL DEFAULT 0,
    accuracy REAL NOT NULL DEFAULT 0,
    duration_seconds INTEGER NOT NULL,
    correct_attempt_timestamps TEXT DEFAULT '[]',
    auto_voice BOOLEAN DEFAULT FALSE,
    used_words TEXT,
    word_definitions TEXT,
    current_word TEXT,
    current_scrambled_word TEXT,
    current_word_definition TEXT,
    current_word_token TEXT,
    used_item_ids TEXT
);

-- Index for phase and score (used for leaderboard queries)
CREATE INDEX IF NOT EXISTS idx_phase_score ON session_entities(phase, score);

-- Index for created_at (used for time-based queries)
CREATE INDEX IF NOT EXISTS idx_created_at ON session_entities(created_at);
