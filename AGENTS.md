## A Test of Your Reflexes

A Test of Your Reflexes, or Atoyr, is a game where the users will see 5 characters in a random order and they will need to re-order those into a correct word, given a definition of the word. For the specs, see the `.opencode/specs` folder for more details. For the skills, see the `.opencode/skills` folder.

## Conventions

- Use TypeScript (except for the `packages/server` which is using Go).
- Use TailwindCSS (v4)
- Use oxfmt for formatting and oxlint for linting
- Use Yarn Modern with nodeLinker node_modules
- Use latest React, no need for useCallback and useMemo
- For static content pages (e.g. `about.tsx`), author in Markdown and render with `marked`. Inject via `dangerouslySetInnerHTML` — the source is a static string in the repo, not user/db-derived content. See `.config/opencode/AGENTS.md` for the policy.
- DO NOT PUT UNNECESSARY COMMENTS between lines unless absolutely necessary. Also don't put unnecessary JSDoc as well for the emitted functions unless the intentions are not clear.
- DO NOT SPLIT INTO MULTIPLE COMPONENTS unless absolutely necessary. If it's possible to colocate the components, co-locate.
- Colocate module-level helper functions at the bottom of the file when they do not close over component state (function declarations are hoisted). Prefer this over nesting them inside the component.
- Bug fixes for features still under active development (same branch/feature) should not be included in the changelog — they are part of the implementation process, not a user-facing change.
- Don't introduce intermediary variables for values used only once or that need no transformation. Read from the source directly.
- When writing specs, put AS LITTLE DETAIL AS POSSIBLE to the implementation details. Just have the higher level; only show code snippets when necessary.
- DO NOT put overly-detailed file structure (apart from top-level ones) because it has potential to change over time.
- DO NOT create additional Markdown files in `.opencode` folder unless otherwise stated.
- Always implement unit/component tests (for UI) and unit/integration tests (for server).
- Always double verify the current implementation and convention in the codebase. Never hallucinate.
- When writing a plan in .opencode/specs, be as detailed as possible.
- ALWAYS ask if you are about to build Docker images. Doing builds alone is okay, but doing builds for Docker images should never be done automatically unless explicitly mentioned.

## Structure

- `packages/client`: The UI, this is the one that the user sees. It communicates with `packages/server` with HTTP and SSE.
- `packages/server`: The server, this is the one that orchestrates the events, words, and stuff.

## Lifecycle

- Specs live in `.opencode/specs` as markdown files describing _what_ to build, not _how_.
- When a spec is implemented, summarize the key decisions and patterns in this file so future sessions have context. The spec file should be removed. Confirm with the user before you remove it.
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

### Healthcheck Endpoint

The server exposes `GET /api/v1/health` (unauthenticated) returning build info and dependency status:

- **Build info**: `name` is hardcoded (`"atoyr"`), `gitHash` is injected via `-ldflags -X main.GitHash=...` at build time.
- **Dependency caching**: `HealthService` (`internal/services/health.service.go`) runs a background goroutine that refreshes every 30s. The handler reads cached values under `sync.RWMutex`.
- **Database check**: calls `db.DB().Ping()`.
- **HTTP status**: 200 if the database is healthy, 500 if the database is unhealthy.
- **Build integration**: `GIT_HASH=$(git rev-parse HEAD)` in `packages/server/Makefile`, passed as ldflags to `build`. Dockerfile uses `ARG GIT_HASH=unknown` with ldflags. CI passes `GIT_HASH=${{ github.sha }}` as a Docker build arg.

Key files: `packages/server/internal/services/health.service.go`, `packages/server/internal/api/health_routes.go`, `packages/server/cmd/service/main.go`.

### Deployment & CI/CD

The project uses GitHub Actions → GHCR (private) → Coolify pull pattern:

- **Why not local builds**: The production VM (4GB/2vCPU) risks OOM with on-server Docker builds. GitHub runners handle compilation instead.
- **GHCR for image registry**: Private images are free on GHCR. `GITHUB_TOKEN` authenticates the push in Actions; a PAT with `read:packages` + `repo` scopes is used for `docker login` on the VM for pulls.
- **Single Dockerfile**: Multi-stage build at project root — node:24-alpine for client SPA, golang:1.25-alpine for Go server, nginx:1.27-alpine as the final runtime. nginx reverse-proxies `/api/` and `/dashboard` to the Go server on port 3000.
- **Deploy workflow** (`.github/workflows/deploy.yml`): Triggered on push to `main`. Builds and pushes to `ghcr.io/<repo>:latest` with GitHub Actions cache (`type=gha`), then curls the Coolify deploy webhook.
- **Key files**: `Dockerfile`, `.dockerignore`, `nginx.conf`, `docker-entrypoint.sh`, `docker-compose.yml`, `.github/workflows/deploy.yml`.

### Client Dependency Notes

A few `packages/client` dependencies are kept even though no application source file imports them directly:

- **`@react-router/node`** and **`isbot`** are required by React Router's framework runtime (`@react-router/dev` / `@react-router/serve`). Removing them breaks `react-router typegen` and the server build. They appear as "unused" in static dependency scans because the framework loads them at runtime rather than through a static import in app code.
- The client package is kept free of dead dependencies otherwise; any future dependency that is only needed transitively by the framework should also be retained as a direct dependency so production installs remain reliable.

### Blind Mode

Blind mode hides the word definition, increasing difficulty:

- **Server-side definition stripping**: The service layer (`game.service.go`) is responsible for omitting the definition — `sessionWithVisibleDefinition()` clears `CurrentWordDefinition` for Blind mode in `StartGame`/`ContinueGame`, and `SubmitAnswer` only sets `ScrambledWordDefinition` when mode is `"vanilla"`. Routes do not filter the definition; the service guarantees it's absent.
- **Mode banner** (`packages/client/app/components/ModeBanner.tsx`): A shared component rendered at the route level in `home.tsx` (right below the navbar). Shows `🚫 BLIND MODE 🚫` or `🍦 Vanilla mode 🍦`.
- **Mode selector**: Lives inside the Settings modal (not a prominent landing-screen control — intentional deviation from the spec).
- **Mode persistence**: Written to localStorage via `packages/client/app/lib/settings.ts` alongside auto-voice. The mode is passed to `POST /api/v1/game/start` as a required field.
- **Leaderboard isolation**: `GetPercentile` now filters by `mode` in SQL so Blind and Vanilla scores are not compared against each other.
- **Speech guard**: `speakLetters` in `GameScreen.tsx` only creates a definition utterance when the `definition` string is non-empty. Since the server omits the definition for Blind mode, no mode-specific branching is needed in the speech code.
- **No analytics events**: Spec'd `ga-blind-mode-started`/`ga-blind-mode-finished` events were skipped because the existing `/dashboard` already tracks session data.
- **No results screen mode indicator**: Skipped because the mode banner during gameplay provides sufficient context.
- **Empty definition contract**: The server returns `scrambledWordDefinition: ""` in blind mode (not omitted). The client must not require this field in conditions gating state updates — `packages/client/app/api/hooks.ts` guards on `scrambledWord && token` only. If `scrambledWordDefinition` is added to the guard, blind mode will silently break (state won't update, `onSuccess` won't fire).

Key files: `packages/server/internal/services/game.service.go`, `packages/client/app/components/ModeBanner.tsx`, `packages/client/app/routes/home.tsx`, `packages/client/app/lib/settings.ts`, `packages/server/internal/services/leaderboard.service.go`.

### Landing & Results Screen Modals

The landing and results screens use a shared `radix-ui` Dialog wrapper for "How to play" and "Settings" modals:

- **Shared wrapper** (`packages/client/app/components/Dialog.tsx`): Owns `Dialog.Root`, `Dialog.Portal`, `Dialog.Overlay`, `Dialog.Content`, close button, and focus-trap/Escape/scroll-lock behavior from `radix-ui`. Consumers pass only the title, trigger, and body content.
- **How to play modal** (`packages/client/app/components/HowToPlayModal.tsx`): Contains the rules list. Trigger lives on the landing screen.
- **Settings modal** (`packages/client/app/components/SettingsModal.tsx`): Contains the auto-voice toggle and its footnote. Triggers live on both the landing screen and the results screen next to "Play Again".
- **Auto-voice persistence**: The toggle still reads from and writes to `localStorage` via `packages/client/app/lib/auto-voice.ts`. `StartScreen` and `useGame.playAgain` read the stored value at action time so the preference is available even when the modal is closed.
- **Landing screen cleanup**: The inline rules list, inline auto-voice checkbox, and footnote were removed from `StartScreen` to keep the landing screen compact. The "Start Game" button remains the primary call to action.
- **Results screen layout**: The settings trigger sits alongside "Back to home" and "Play Again".

Key files: `packages/client/app/components/Dialog.tsx`, `packages/client/app/components/HowToPlayModal.tsx`, `packages/client/app/components/SettingsModal.tsx`, `packages/client/app/components/StartScreen.tsx`, `packages/client/app/components/ResultsScreen.tsx`.

## Tests

Run top level `yarn test` to run all tests in all packages. Otherwise, use `yarn workspaces <folder_name>` to run individual tests. If possible, ALWAYS add unit tests with `vitest` for any logic-related functionalities. For UI related functionalities (such as CSS), it is not necessary unless otherwise stated.

### Test Query Conventions

When writing tests, prefer queries in this order:

1. **`getByRole`** — for interactive elements (`button`, `textbox`) and landmark/heading roles (`heading`, `section`). Use this whenever the element has an implicit or explicit role.
2. **`within(section).getByText(...)`** — for content scoped to a semantic parent. Always scope queries to the relevant `<section>` or container (e.g., `within(getByRole('heading', { name: 'Score' }).closest('section')!).getByText('2/4 correct')`).
3. **`getByTestId`** — only as a last resort when no semantic query is possible.

Avoid bare `getByText` for stat values or labels that belong to a specific region — scope them with `within()` instead.

For elements whose visible text is split across siblings (e.g. an `sr-only` label + a visible value rendered in the same slot), prefer a single `getByText((_, node) => ...)` custom matcher over two separate assertions on the label and the value. The matcher should (a) compare `node.textContent` to the combined text and (b) constrain the node structurally so the matcher doesn't also match ancestor containers — e.g. `node.parentElement?.dataset.testid === 'answer-slots'`. One assertion that captures intent ("the first slot renders 'P'") is better than two that recheck the same DOM node.

### Avoid Testing Implementation Details in Component Tests

Component tests should verify user-observable behavior, not internal attributes or wiring. For example, do **not** assert on `data-ga-label`, `data-testid` values, CSS classes, or other implementation-specific hooks in component tests. If an analytics label or tracking contract needs regression coverage, test it in a dedicated analytics/tracking test or an integration test instead. See `packages/client/app/components/ResultsScreen.test.tsx` as an example of what not to do.

### Route-Level Tests: Assert Transitions, Not Rendering

Route-level tests (e.g., `home.test.tsx`) orchestrate screen transitions. They should assert on **which screen is visible**, not on rendering details of the child component. Component-level tests (`StartScreen.test.tsx`, `GameScreen.test.tsx`) own the rendering assertions.

- **Do**: assert on transition evidence — a heading or button unique to the target screen (e.g., `Remaining seconds` heading proves the playing screen rendered).
- **Don't**: assert on specific values rendered by the child component (e.g., timer `"30s"`, score `"0/0 correct"`, definition text, or scrambled word). Those belong in the component's own tests.
- **Why**: route test failures should indicate orchestration bugs (wrong screen shown, state not reset), not rendering bugs in a child component. This keeps test failures scoped and reduces brittle assertions.

### Changelog Page & Version Indicator

The navbar shows a combined version chip (client + server semver added component-wise) instead of the `?` icon:

- **Version calculation** (`packages/client/app/lib/version.ts`): `computeCombinedVersion` parses `import.meta.env.VERSION` (format `client-server`), adds each semver component arithmetically, and returns `v{major}.{minor}.{patch}`. Falls back to the raw `VERSION` string if parsing fails.
- **Version chip** (`packages/client/app/components/VersionHoverCard.tsx`): Uses Radix `HoverCard`. Hover shows client/server version breakdown and a link to `/changelog`. Opening the hover panel dismisses the "New" badge and writes `lastSeen`/`dismissedAt` to localStorage.
- **"New" badge**: Shown when `lastSeen !== currentVersion`, OR when `lastSeen === currentVersion` but the player hasn't dismissed the badge yet and the current date is at least one week after the latest release date.
- **Changelog page** (`packages/client/app/routes/changelog.tsx`): Renders markdown from `packages/client/app/lib/changelog.ts` using `marked`, following the same pattern as `about.tsx`.
- **Changelog markdown** (`packages/client/app/lib/changelog.ts`): Uses `# Month Year` for h1 and `## Week N` (N=1-5) for h2. The changelog is ordered newest-first (descending by month/week). The first h1/h2 pair derives the latest release date for the badge logic. Changelog entries must be user-facing and essential. Do not log internal refactors or dependency-only changes. Do not use em dash (—) in changelog entries; use ";" instead.
- **Navbar hover**: All navbar links (`Play`, `Leaderboard`, `About`) have `hover:text-dark-interactive-primary transition duration-200` for consistent hover feedback.

Key files: `packages/client/app/components/VersionHoverCard.tsx`, `packages/client/app/lib/version.ts`, `packages/client/app/lib/changelog.ts`, `packages/client/app/routes/changelog.tsx`, `packages/client/app/components/PageLayout.tsx`, `CHANGELOG.md`.
