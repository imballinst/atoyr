ALTER TABLE session_entities ADD COLUMN word_definitions TEXT;

DROP INDEX IF EXISTS idx_topic_mode_phase_score_accuracy;

CREATE INDEX IF NOT EXISTS idx_mode_phase_score_accuracy ON session_entities(mode, phase, score, accuracy);

ALTER TABLE session_entities DROP COLUMN topic;
