-- 001_initial_schema.down.sql
-- Rollback initial schema

DROP INDEX IF EXISTS idx_created_at;
DROP INDEX IF EXISTS idx_phase_score;
DROP TABLE IF EXISTS session_entities;
