# Backend migration to AGS

> Status: spec / design

## Goal

Document which backend functionalities of Atoyr are supported and migrate-able to AccelByte Gaming Services (AGS), and which ones are better left in the custom Go server or handled through AGS Extend.

## Scope

Covers `packages/server` only: game session lifecycle, scoring, leaderboard, admin dashboard, observability, and authentication.

## Current backend capabilities

- Anonymous game sessions via a cookie-based `session_id`.
- Single-player word-scramble game loop: start, continue, submit answer, finish, and timer expiry.
- In-memory countdown timer with an SSE tick stream.
- SQLite persistence for round state, score, attempts, accuracy, and used words.
- Mode-filtered global leaderboard and percentile ranking.
- Admin dashboard with Google OAuth, real-time stats, and time-series request metrics.
- Health endpoint and Sentry error reporting.

## AGS migration map

| Current capability | AGS target | Migrate? | Notes |
|---|---|---|---|
| Anonymous player identity | IAM headless account / OAuth | Yes | Replace the cookie-only `session_id` with an AGS IAM token so sessions belong to a real player identity. A short-lived round cookie can still identify the active round. |
| Admin dashboard authentication | IAM OAuth + roles/permissions | Yes | Replace Google OAuth with an AGS IAM admin client and gate `/admin/*` through AGS roles/permissions. |
| Final score, attempts, and accuracy per player | Statistics module | Yes | Post round results as server-authoritative statistics, per mode. Use stat codes such as `atoyr-score-vanilla` and `atoyr-score-blind`. Accuracy can be stored as additional data or a separate stat. |
| Global leaderboard and percentile | Leaderboards module | Yes* | Create mode-specific leaderboards backed by the mode statistics. Native AGS ranking is by a single stat; tie-breaking on accuracy/end time requires an Extend override or a composite scoring formula. Percentile can be derived from AGS ranked lookup and total entries. |
| Gameplay events (round start, answer submit, round finish) | Analytics module | Yes | Fire custom AGS analytics events for funnel, retention, and engagement reporting. Operational HTTP metrics stay outside AGS Analytics. |
| Round items / power-ups (`itemsUsed`) | Store/Entitlements + Achievements | Optional future | If `itemsUsed` becomes real consumables or unlockables, model them in AGS Store/Entitlements and gate usage through Achievements or entitlements. |
| Round-specific game state (used words, current token, timer) | Cloud Save, or keep custom | Partially | This state is short-lived and gameplay-specific. If resume-across-sessions is needed, store an opaque save slot in Cloud Save. Otherwise keep it in the custom server. |
| Countdown timer and SSE tick stream | none native | No | Keep the custom timer and SSE endpoint. AGS Session Management is for multiplayer match/session wrappers, not single-player countdowns. |
| Word bank and scrambling logic | none native | No | Keep custom; the word list is static repo content. |
| Request-level metrics, latency percentiles, and memory snapshots | none native | No | Keep the custom dashboard or replace with external observability. AGS Analytics is player-behavior telemetry, not server performance monitoring. |
| Health endpoint, CORS, graceful shutdown, Sentry | none | No | These are infrastructure concerns and remain custom. |

## Per-endpoint migration behavior

### `POST /api/v1/game/start`

Before:
- Validates `mode` and `itemsUsed`.
- Creates a `SessionEntity` row in SQLite in the `idle` phase.
- Picks a word, scrambles it, generates a token, sets the phase to `playing`, computes `EndsAt`, and starts an in-memory countdown.
- Sets a `session_id` cookie.
- Returns the scrambled word, definition (or empty in blind mode), token, and remaining seconds.

After:
- Validates the request.
- Ensures the caller has an AGS IAM token; if not, creates or reuses a headless account bound to a randomized device ID and returns the IAM access token.
- Creates a Cloud Save game record keyed by round ID, owned by that user, storing the initial round state: mode, autoVoice, itemsUsed, usedWords, currentWord, currentScrambledWord, currentWordDefinition, currentWordToken, score, totalAttempts, correctAttemptTimestamps, and endsAt.
- Initializes the same custom in-memory countdown timer.
- Sets a short-lived round cookie (e.g. `atoyr_round_id`) alongside the IAM token.
- Returns the same response shape as today.

### `POST /api/v1/game/continue`

Before:
- Reads the `session_id` cookie and loads the SQLite row.
- Resumes the in-memory timer if the round is still playing.
- Returns the current round state.

After:
- Reads the IAM token to identify the player and the round cookie to identify the Cloud Save record.
- Fetches the round state from Cloud Save.
- Resumes the in-memory timer.
- Returns the current round state.

### `POST /api/v1/game/submit`

Before:
- Loads the SQLite row by cookie.
- Validates the token and answer, updates score/attempts/accuracy, picks the next word when correct, applies the time penalty or bonus, and writes everything back to SQLite.
- Updates the in-memory countdown.
- Returns the result.

After:
- Identifies the player via IAM token and the round via the round cookie.
- Fetches the round state from Cloud Save.
- Runs the same custom validation, scoring, and word-selection logic.
- Writes the updated state back to Cloud Save.
- Updates the in-memory countdown.
- Returns the same response shape as today.

### Timer expiry and `FinishGame`

Before:
- Marks the SQLite session as finished, writes the final score, accuracy, and end time, removes the in-memory timer entry, and makes the row eligible for the leaderboard.

After:
- Marks the Cloud Save record as finished.
- Posts the final `score`, `totalAttempts`, and `accuracy` to AGS Statistics under mode-specific stat codes.
- AGS Leaderboards update rankings automatically from the statistics.
- Removes the in-memory timer entry.

### `GET /api/v1/game/sse`

Before:
- Validates the `session_id` cookie, reads the SQLite row, and streams tick events from the in-memory timer until the round finishes.

After:
- Validates the round cookie and IAM token, reads the round state from Cloud Save when needed, and streams tick events from the same in-memory timer.

### `GET /api/v1/leaderboard`

Before:
- Queries SQLite for finished sessions filtered by mode, orders by score, accuracy, and end time, and paginates the result.

After:
- Calls the AGS Leaderboards top-N query for the requested mode leaderboard, which is backed by the mode-specific statistic.
- Transforms AGS rank entries into the existing response shape.
- Total count comes from AGS leaderboard metadata or a count query.

### `GET /api/v1/leaderboard/percentile`

Before:
- Loads the current session from SQLite, confirms it is finished, and computes the percentile using the custom tie-breaking logic.

After:
- Loads the player's finished round from Cloud Save.
- Uses AGS Leaderboards ranked lookup for the player's stat value plus the total entry count to derive a percentile.
- If the custom tie-breaker must be preserved, keep a custom fallback or use an AGS Extend Override.

### `GET /api/v1/health`

Before/after: unchanged. The endpoint remains a custom health check for the Atoyr server and its database connection.

### `GET /api/v1/admin/stats`

Before:
- Queries SQLite `session_entities` and `metrics_snapshots` for session counts, active games, and mode breakdown.

After:
- Session and engagement counts come from AGS Analytics.
- Active games may still be counted from the custom server's in-memory timer registry.
- Operational request/error counts remain custom.

### `GET /api/v1/admin/timeseries`

Before:
- Reads `metrics_snapshots` from SQLite.

After:
- Gameplay event time series come from AGS Analytics.
- Operational request-latency, error-rate, and memory-usage time series remain custom.

## Migration shape

### Phase 1: Identity and stats

- Introduce AGS IAM login (headless or OAuth) before a round starts.
- Replace the cookie-based `session_id` as the sole identity with an AGS player token, while keeping a round cookie for the active game session.
- At round end, post the final score and attempts as server-authoritative AGS statistics.

### Phase 2: Leaderboards

- Configure AGS Leaderboards per mode, backed by the statistics from Phase 1.
- Decide whether to keep the custom tie-breaking rule (score → accuracy → end time) or replace it with a single AGS stat, such as a composite score.
- If the tie-breaker must stay, implement it through an AGS Extend Override or keep the custom leaderboard service.
- Surface AGS top-N and around-me queries instead of the custom SQLite leaderboard.

### Phase 3: Analytics

- Fire custom AGS analytics events for `round_started`, `answer_submitted`, and `round_finished`.
- Use AGS dashboards for DAU, retention, and engagement; keep the custom dashboard for request-level operational metrics.

### Phase 4: Items and entitlements (optional)

- If items become real game features, model them in AGS Store/Entitlements and wire unlock checks into the round start flow.

## What stays in the custom server

- The word bank, scrambling, and answer validation logic.
- The in-memory countdown timer and SSE tick stream.
- Round-state persistence (unless Cloud Save is adopted for cross-session resume).
- Request-level metrics, health endpoint, CORS, Sentry, and graceful shutdown.

## Risks and open questions

- **Anonymous vs. authenticated**: moving from a fully anonymous cookie to AGS IAM changes the landing-screen UX and may require a login or silent headless flow.
- **Leaderboard tie-breaking**: AGS native leaderboards rank by a single statistic. Preserving the current multi-field tie-breaker is the biggest functional gap and likely requires Extend or a composite score.
- **Per-round state**: AGS Session Management is designed for multiplayer match wrappers, not single-player puzzle rounds. Storing round state in Cloud Save is possible but adds latency; keeping it in the custom server is simpler.
- **Operational dashboard**: AGS does not replace the existing request-latency and error-rate dashboard. Decide whether to keep that dashboard or migrate it to a separate observability stack.
