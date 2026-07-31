ALTER TABLE session_entities ADD COLUMN topic TEXT NOT NULL DEFAULT 'english-words';

DROP INDEX IF EXISTS idx_mode_phase_score_accuracy;

CREATE INDEX IF NOT EXISTS idx_topic_mode_phase_score_accuracy ON session_entities(topic, mode, phase, score, accuracy);

ALTER TABLE session_entities DROP COLUMN word_definitions;
