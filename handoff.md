# AGS Migration Plan

## Motivation

Remove the SQLite dependency from the game server entirely. AGS serves as both the operational store (via in-memory state + async AGS persistence) and the analytics/leaderboard backend.

## Trade-off

In-memory state means in-flight sessions are lost on server crash. Acceptable for this game — timer restoration from SQLite is the only remaining SQLite read after migration.

## Architecture

### 1. Operational State: In-memory + Async AGS Persistence

Replace synchronous SQLite `FindByID`/`Update` calls with an in-memory session store:

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Session store | `sync.Map` keyed by `sessionID` | Sub-millisecond reads/writes during gameplay |
| Timer store | `SessionDurationManager` (already exists) | In-memory countdown per session |
| Persistence queue | Goroutine + channel | Flush session state to AGS Cloud Save asynchronously |
| Crash recovery | SQLite (only remaining use) | `SELECT ... WHERE phase = 'playing'` on startup to rebuild in-memory state + timers |

**Flow per game:**
1. `StartGame` → insert into `sync.Map`, persist to AGS Cloud Save (async)
2. `SubmitAnswer` → read/write `sync.Map`, persist to AGS Cloud Save (async)
3. `FinishGame` → read from `sync.Map`, publish stats to AGS User Stats + Leaderboard (async), delete from `sync.Map`

### 2. Leaderboard & Percentile: Always SQLite (already fixed)

The leaderboard fix from the previous session (removing AGS delegation from `GetLeaderboard`, `GetTotalEntries`, `GetPercentile`) stands. These always read from SQLite, which is written synchronously and has all fields (`totalAttempts`, `accuracy`, `timestamp`, session `ID`).

AGS stat posting (`PostRoundStats`) continues to run asynchronously — it's a side effect, not the source of truth for leaderboard queries.

### 3. Dashboard Stats: SQLite Polling

The stats service (`GET /admin/stats`, `GET /admin/timeseries`) stays on SQLite. AGS Analytics can ingest session-started/finished custom events for export to an external warehouse (S3/Redshift/Snowflake) if needed later, but the real-time dashboard queries the local DB.

### 4. Composite Score (Already Done)

`calculateCompositeScore` (`score * 1,000,000 + round(accuracy * 10,000) - durationSeconds`) encodes all three tiebreak dimensions into a single sortable integer. AGS leaderboard and percentile ranking use this composite.

### 5. Cloud Save: Key-Value Only

AGS Cloud Save is a key-value store per user (`{userID, key}` → value). No field-level querying. The session ID is used as the key; scanning active sessions on startup iterates the known set from SQLite (see Crash Recovery above).

### 6. Analytics: Custom Events + External Export

Session-started/finished events emit via Analytics module custom telemetry. Export pipeline (S3/Redshift/Snowflake) is natively supported; SQLite polling is the simpler path for the current dashboard.

## Remaining SQLite Footprint After Migration

| Query | Replaced by |
|-------|------------|
| `FindByID` / `Update` during gameplay | In-memory `sync.Map` + async AGS Cloud Save |
| `ORDER BY score DESC, accuracy DESC, ends_at ASC` (leaderboard) | Stays on SQLite |
| `GetPercentile` percentile computation | Stays on SQLite |
| `SELECT ... WHERE phase = ? AND mode = ? AND score > 0` (total entries) | Stays on SQLite |
| `COUNT(*) ... WHERE created_at >= ?` (stats) | Stays on SQLite (or Analytics export) |
| `SELECT ... WHERE phase = 'playing'` (crash recovery) | Stays on SQLite (only write after migration) |
