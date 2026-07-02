-- 002_add_metrics_snapshots.sql
-- Metrics snapshots table for time series data

CREATE TABLE metrics_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL,
    
    -- Request metrics (delta since last snapshot)
    request_count INTEGER NOT NULL DEFAULT 0,
    error_count_4xx INTEGER NOT NULL DEFAULT 0,
    error_count_5xx INTEGER NOT NULL DEFAULT 0,
    
    -- Game metrics
    active_games INTEGER NOT NULL DEFAULT 0,
    
    -- Response time metrics (in microseconds for precision)
    response_time_p50 INTEGER NOT NULL DEFAULT 0,
    response_time_p95 INTEGER NOT NULL DEFAULT 0,
    response_time_p99 INTEGER NOT NULL DEFAULT 0,
    
    -- Server resources
    memory_usage_mb REAL NOT NULL DEFAULT 0
);

-- Index for fast time-range queries
CREATE INDEX idx_metrics_snapshots_timestamp ON metrics_snapshots(timestamp);
