## A Test of Your Reflexes

A Test of Your Reflexes, or Atoyr, is a game where the users will see 5 characters in a random order and they will need to re-order those into a correct word, given a definition of the word. For the specs, see the `.opencode/specs` folder for more details. For the skills, see the `.opencode/skills` folder.

## Conventions

- Use TypeScript (except for the `packages/server` which is using Go).
- Use TailwindCSS (v4)
- Use oxfmt for formatting and oxlint for linting
- Use Yarn Modern with nodeLinker node_modules
- Use latest React, no need for useCallback and useMemo
- DO NOT PUT UNNECESSARY COMMENTS between lines unless absolutely necessary. Also don't put unnecessary JSDoc as well for the emitted functions unless the intentions are not clear.
- DO NOT SPLIT INTO MULTIPLE COMPONENTS unless absolutely necessary. If it's possible to colocate the components, co-locate.
- When writing specs, put AS LITTLE DETAIL AS POSSIBLE to the implementation details. Just have the higher level; only show code snippets when necessary.
- DO NOT put overly-detailed file structure (apart from top-level ones) because it has potential to change over time.
- DO NOT create additional Markdown files in `.opencode` folder unless otherwise stated.
- Always implement unit/component tests (for UI) and unit/integration tests (for server).
- Always double verify the current implementation and convention in the codebase. Never hallucinate.
- When writing a plan in .opencode/specs, be as detailed as possible.

## Structure

- `packages/client`: The UI, this is the one that the user sees. It communicates with `packages/server` with HTTP and SSE.
- `packages/server`: The server, this is the one that orchestrates the events, words, and stuff.

## Lifecycle

- Specs live in `.opencode/specs` as markdown files describing *what* to build, not *how*.
- When a spec is implemented, summarize the key decisions and patterns in this file so future sessions have context. The spec file should be removed.
- Keep this file updated as the source of truth for project conventions and architecture.

## Implemented Architecture Decisions

### Tick Restoration and Graceful Restart

The server handles container restarts (Coolify deploys) gracefully:

- **Penalty persistence**: Wrong-answer penalty is written to `EndsAt` in SQLite (not just in-memory), so timers survive container restarts.
- **Restorable timers**: `SessionDurationManager` entries are reconstructed on startup by scanning for `phase = "playing"` sessions and computing remaining time from `EndsAt`.
- **Graceful shutdown**: The server handles `SIGTERM` and drains active connections before exiting (45s timeout).
- **SSE reconnection**: The client automatically reconnects on SSE disconnection after a short delay.

Key files: `packages/server/cmd/main.go`, `packages/server/internal/services/game.service.go`, `packages/server/internal/services/session.service.go`.

### Server Monitoring Dashboard

The server has a monitoring dashboard at `/dashboard` with Google OAuth authentication:

- **In-memory metrics collector** (`internal/middleware/metrics.go`): Atomic counters for request count, error counts (4xx/5xx), active games, and a ring buffer (cap 1000) for response time percentiles (P50/P95/P99).
- **SQL migrations** (`migrations/`, `internal/database/migrations.go`): Uses `golang-migrate/v4` instead of GORM AutoMigrate. Numbered `.up.sql`/`.down.sql` files with a `schema_migrations` version-tracking table.
- **Stats service** (`internal/services/stats.service.go`): Two API endpoints -- `GET /admin/stats` (real-time metrics from memory + session counts from SQLite) and `GET /admin/timeseries?period=&granularity=` (historical data from `metrics_snapshots` table).
- **Snapshot goroutine**: Background goroutine in `MetricsCollector.StartSnapshotWorker` writes delta-based snapshots to SQLite every minute for historical time series.
- **Google OAuth** (`internal/middleware/auth.go`): Uses `go-oidc` + `oauth2` to verify Google ID tokens. Checks email against `ALLOWED_ADMIN_EMAILS` env var. Includes `NewNoopAuthMiddleware()` for local dev without Google auth.
- **Dashboard** (`web/static/dashboard.html`): Vanilla HTML with TailwindCSS, Chart.js, and Google Identity Services via CDN. Polls `/admin/stats` every 5s. Shows 4 stat cards + 4 time series charts with period selector (1h/24h/7d/1M).
- **Admin routes** (`internal/api/admin_routes.go`): Registered under `/admin` group with `RequireAuth()` middleware. Routes: `GET /admin/stats`, `GET /admin/timeseries`.
- **Stats service has a `Now` field** for testability (time mocking).
- **Test** (`internal/services/stats_test.go`): Unit tests for GetStats (with/without data) and GetTimeSeries (normal, empty, invalid period).

Key files: `packages/server/cmd/main.go`, `packages/server/internal/services/game.service.go`, `packages/server/internal/services/session.service.go`, `packages/server/internal/middleware/metrics.go`, `packages/server/internal/middleware/auth.go`, `packages/server/internal/services/stats.service.go`, `packages/server/internal/api/admin_routes.go`, `packages/server/web/static/dashboard.html`.

## Tests

Run top level `yarn test` to run all tests in all packages. Otherwise, use `yarn workspaces <folder_name>` to run individual tests. If possible, ALWAYS add unit tests with `vitest` for any logic-related functionalities. For UI related functionalities (such as CSS), it is not necessary unless otherwise stated.
