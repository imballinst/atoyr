# Database Decision: SQLite vs Postgres

**Status:** Decision recorded — SQLite remains the default for now.

## Context

Atoyr currently uses SQLite as the primary database. The server is a single Go binary running in one container (Coolify), with a 4GB/2vCPU production VM. The workload is session-based: per-session reads, occasional writes to `EndsAt` for penalty persistence, and periodic metrics snapshots.

## Decision

Keep SQLite until one of the migration triggers below is reached. Do not migrate preemptively.

## Pros and cons in this context

### SQLite

**Pros**
- Single process/container. No extra database service competing for limited VM resources.
- Simple deployment and local development. Tests can use `:memory:` or temp files.
- Whole-database backup is a single file copy.
- Good enough for the current workload: mostly reads and infrequent writes per active session.

**Cons**
- Single-writer lock. Concurrent write spikes (e.g., many sessions ending at once) can queue and cause latency.
- Hard to scale horizontally. Two server containers cannot safely share a SQLite file.
- Fewer operational tools for observability, backups, and long-running analytics.

### Postgres

**Pros**
- Better concurrent write performance.
- Enables horizontal scaling by sharing one database across multiple server containers.
- Richer querying, indexing, and analytics for reporting over `metrics_snapshots` and player data.
- Mature backup, monitoring, and high-availability ecosystem.

**Cons**
- Adds operational overhead: another container/service, backups, updates, connection management.
- Consumes resources on an already small 4GB VM.
- More complex local development setup.
- Requires migration engineering time and ongoing maintenance.

## Concurrency thresholds

The number that matters most is **sustained writes per second**, not just active sessions. With WAL mode enabled, SQLite handles many readers in parallel, but only one writer at a time.

### Rough rules of thumb for Atoyr

| Zone | Concurrent active sessions | Sustained writes/sec | Action |
|------|---------------------------|---------------------|--------|
| Green | < 500 | < 50 | SQLite is comfortable. |
| Yellow | 500–2,000 | 50–200 | Monitor P99 latency. Optimize before migrating. |
| Red | > 2,000 | > 200 | Strongly consider Postgres, or if you need horizontal scaling. |

### Why these numbers

- A typical session writes only a handful of times per round (start, answer, end, penalty update). So 1,000 active sessions may produce only 20–50 writes/sec, not 1,000.
- The bigger risk is **write bursts**: many sessions ending at the same time, or the minute-ly metrics snapshot landing while other writes queue up. Short bursts are fine; sustained high write rates are not.
- SQLite with WAL can theoretically do hundreds of writes/sec, but in practice with fsync, connection overhead, and contention, 50–200 sustained writes/sec is a realistic comfort zone.

### Important caveats

- WAL mode must be enabled. Without it, readers and writers block each other and the ceiling is much lower.
- Long transactions lower the ceiling. Keep writes small and fast.
- The real signal is application latency, not user count. If P99 stays healthy, the number of sessions is less important.
- These are estimates, not guarantees. Measure with production-like load before deciding.

## Migration triggers

Migrate to Postgres when any of the following becomes true:

1. Sustained write throughput is regularly above ~200 writes/sec, or P95/P99 latency spikes during round ends and snapshot writes become unacceptable.
2. The app needs more than one server container.
3. Analytics or reporting requirements outgrow SQLite’s query and indexing capabilities.
4. Operational requirements demand automated backups, point-in-time recovery, or replicas.
5. Data volume or throughput exceeds comfortable SQLite limits.

## Migration cost

- **Schema migration:** Existing `golang-migrate/v4` setup transfers, but SQL dialects differ (integer autoincrement, date functions, locking).
- **Code changes:** Review raw SQL/GORM usage, especially in `game.service.go` and `session.service.go` around `EndsAt` and timer restoration.
- **Deployment changes:** Add a Postgres container or managed DB, connection string via env, possible connection pooling.
- **Operations:** Own backup automation, monitoring, disk growth, and version upgrades.

## Bottom line

SQLite is the right choice while Atoyr remains a single-container app with moderate concurrent load. Postgres becomes the better option once the app needs horizontal scaling or measurable write contention appears.
