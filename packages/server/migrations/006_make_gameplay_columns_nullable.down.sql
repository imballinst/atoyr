-- 006_make_gameplay_columns_nullable.down.sql

ALTER TABLE session_entities RENAME COLUMN duration_seconds TO duration_seconds_old;
ALTER TABLE session_entities ADD COLUMN duration_seconds INTEGER NOT NULL DEFAULT 0;
UPDATE session_entities SET duration_seconds = COALESCE(duration_seconds_old, 0);
ALTER TABLE session_entities DROP COLUMN duration_seconds_old;
