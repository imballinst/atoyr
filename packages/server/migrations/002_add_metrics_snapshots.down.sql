-- 002_add_metrics_snapshots.down.sql
-- Rollback metrics snapshots table

DROP INDEX IF EXISTS idx_metrics_snapshots_timestamp;
DROP TABLE IF EXISTS metrics_snapshots;
