DROP INDEX IF EXISTS idx_phase_score_accuracy;

CREATE INDEX IF NOT EXISTS idx_phase_score ON session_entities(phase, score);
