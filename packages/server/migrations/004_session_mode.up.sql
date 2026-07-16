-- Column.
ALTER TABLE session_entities ADD COLUMN mode TEXT NOT NULL DEFAULT "vanilla";

-- Index.
DROP INDEX IF EXISTS idx_phase_score_accuracy;

CREATE INDEX IF NOT EXISTS idx_mode_phase_score_accuracy ON session_entities(mode, phase, score, accuracy);
