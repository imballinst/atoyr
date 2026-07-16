-- Index.
DROP INDEX IF EXISTS idx_mode_phase_score_accuracy;

CREATE INDEX IF NOT EXISTS idx_phase_score_accuracy ON session_entities(phase, score, accuracy);

-- Table column.
ALTER TABLE session_entities DROP COLUMN mode;
