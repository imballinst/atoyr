-- Index for phase, score, and accuracy (used for leaderboard queries)
DROP INDEX IF EXISTS idx_phase_score;

CREATE INDEX IF NOT EXISTS idx_phase_score_accuracy ON session_entities(phase, score, accuracy);
