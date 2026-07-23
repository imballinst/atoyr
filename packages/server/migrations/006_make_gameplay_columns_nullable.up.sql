-- 006_make_gameplay_columns_nullable.sql
-- The minimal session registry no longer writes gameplay columns during active
-- gameplay. Make duration_seconds nullable with a default so that registry
-- inserts do not violate the old NOT NULL constraint.

ALTER TABLE session_entities RENAME COLUMN duration_seconds TO duration_seconds_old;
ALTER TABLE session_entities ADD COLUMN duration_seconds INTEGER DEFAULT 0;
UPDATE session_entities SET duration_seconds = COALESCE(duration_seconds_old, 0);
ALTER TABLE session_entities DROP COLUMN duration_seconds_old;
