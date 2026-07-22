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
| Global leaderboard and percentile | Leaderboards module | Yes* | Create mode-specific leaderboards backed by the mode statistics. AGS Leaderboards ranks players, not individual sessions, so the product shifts from one entry per round to one entry per player. Native ranking is by a single stat; tie-breaking on accuracy/end time requires an Extend override or a composite scoring formula. Percentile can be derived from AGS ranked lookup and total entries. |
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
- Fetches the round state from the in-memory cache.
- Runs the same custom validation, scoring, and word-selection logic.
- Updates the in-memory state.
- Writes the updated state to Cloud Save asynchronously (fire-and-forget) so a restart can recover the round; the response does not wait on this write.
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
- Validates the round cookie and IAM token and streams tick events from the in-memory timer. Cloud Save is not consulted during the stream; recovery happens when the client reconnects and calls ContinueGame.

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
- Operational counts (sessions today, active games, mode breakdown) are counted from the custom server's minimal round registry or in-memory timer state.
- AGS Analytics provides out-of-box dashboards (DAU, MAU, PCCU) and CSV export for engagement metrics; it does not provide a real-time queryable API that can replace these endpoints.

### `GET /api/v1/admin/timeseries`

Before:
- Reads `metrics_snapshots` from SQLite.

After:
- Operational time series (request count, error rate, latency percentiles, memory usage) remain custom.
- Gameplay event time series are available through AGS Analytics dashboards or warehouse export, not through a real-time query API on this endpoint.

## Schema migration

After the migration, the SQLite `session_entities` table is replaced by two concerns:

1. A **round registry** in SQLite for deploy recovery and admin counting.
2. **Round state** in AGS Cloud Save.

### Fields to remove from `session_entities`

These move to AGS Cloud Save and/or AGS Statistics:

- `Score` → AGS Statistics (mode-specific stat code).
- `TotalAttempts` → AGS Statistics.
- `Accuracy` → AGS Statistics or computed from Statistics on read.
- `DurationSeconds` → derived from `EndsAt` during recovery.
- `CorrectAttemptTimestamps` → Cloud Save round state.
- `AutoVoice` → Cloud Save round state.
- `UsedWords` → Cloud Save round state.
- `WordDefinitions` → unused; drop.
- `CurrentWord` → Cloud Save round state.
- `CurrentScrambledWord` → Cloud Save round state.
- `CurrentWordDefinition` → Cloud Save round state.
- `CurrentWordToken` → Cloud Save round state.
- `UsedItemIDs` → Cloud Save round state.

### Fields to keep in the minimal round registry

- `ID` — round identifier; also the Cloud Save record key.
- `UserId` (new) — the AGS player ID tied to the round.
- `Mode` — needed for admin breakdown and recovery.
- `Phase` — needed for startup recovery queries.
- `CreatedAt` — needed for sessions today/week/month counts.
- `EndsAt` — needed for timer recovery.
- `UpdatedAt` — optional DB metadata.

### Recovery flow

On startup, the server queries the registry for `phase = 'playing'` rounds, loads each round's full state from Cloud Save, and restores the in-memory timer. This replaces the current `restoreActiveSessions` scan of the full SQLite table.

### Optional: remove the table entirely

If the admin dashboard and deploy recovery are redesigned to not need a local queryable registry, the SQLite table can be dropped completely. In practice, Cloud Save does not support efficient cross-user queries, so a minimal local registry is recommended.

## AGS SDK configuration and backend data structures

This migration uses the server-side AccelByte Go SDK (`accelbyte-go-sdk`) for IAM, Cloud Save, Statistics, Leaderboards, and Analytics calls from the custom Go server. AGS Extend is only needed if we later introduce custom backend logic that AGS native services cannot support, such as the current multi-field tie-breaker.

### SDK configuration

The server initializes one `accelbyte-go-sdk` `Config` from environment variables:

- `AGS_BASE_URL` — Shared Cloud or Private Cloud endpoint.
- `AGS_NAMESPACE` — the game namespace.
- `AGS_CLIENT_ID` / `AGS_CLIENT_SECRET` — server-to-server OAuth client credentials.
- `AGS_ADMIN_REDIRECT_URI` — callback URI for admin dashboard OAuth flows.

The SDK handles access-token caching and refresh; the server stores only configuration on disk.

### Identity tokens

Three token paths are needed:

1. **Player token** — issued via IAM headless account or platform OAuth, sent by the client in the `Authorization` header or a secure cookie.
2. **Admin token** — issued via IAM OAuth for an admin client, gated by AGS roles/permissions.
3. **Server token** — client-credentials token used by `accelbyte-go-sdk` for service-to-service calls.

The custom server validates player and admin tokens by calling AGS IAM introspection or by using SDK helpers.

### Backend config struct

```go
type AGSConfig struct {
    BaseURL      string
    Namespace    string
    ClientID     string
    ClientSecret string
    RedirectURI  string
}
```

### Cloud Save round state

Each active round is stored as a Cloud Save record owned by the player. The record key is deterministic, e.g. `atoyr:round:{roundID}`. Tags can include `mode:{vanilla|blind}` and `phase:{idle|playing|finished}`.

Record payload:

```go
type CloudSaveRoundState struct {
    Mode                     string   `json:"mode"`
    Phase                    string   `json:"phase"`
    AutoVoice                bool     `json:"autoVoice"`
    ItemsUsed                []string `json:"itemsUsed"`
    UsedWords                []string `json:"usedWords"`
    CurrentWord              string   `json:"currentWord"`
    CurrentScrambledWord     string   `json:"currentScrambledWord"`
    CurrentWordDefinition    string   `json:"currentWordDefinition"`
    CurrentWordToken         string   `json:"currentWordToken"`
    Score                    int32    `json:"score"`
    TotalAttempts            int32    `json:"totalAttempts"`
    Accuracy                 float32  `json:"accuracy"`
    CorrectAttemptTimestamps []int64  `json:"correctAttemptTimestamps"` // Unix ms
    EndsAt                   int64    `json:"endsAt"`                   // Unix ms
    CreatedAt                int64    `json:"createdAt"`                // Unix ms
    UpdatedAt                int64    `json:"updatedAt"`                // Unix ms
}
```

Cloud Save is the source of truth for round state across deploys; the in-memory cache is the source of truth during a single process lifetime.

### Statistics configuration

Create one server-authoritative stat per metric per mode:

| Stat code | Type | Description |
|---|---|---|
| `atoyr_score_vanilla` | integer | Final round score in vanilla mode. |
| `atoyr_score_blind` | integer | Final round score in blind mode. |
| `atoyr_accuracy_vanilla` | float | Final accuracy in vanilla mode. |
| `atoyr_accuracy_blind` | float | Final accuracy in blind mode. |
| `atoyr_attempts_vanilla` | integer | Final attempt count in vanilla mode. |
| `atoyr_attempts_blind` | integer | Final attempt count in blind mode. |

All stats are server-authoritative; the client never writes them. Tags can include `mode` and `season`.

### Leaderboard configuration

Create one leaderboard per mode, backed by the primary score stat:

| Leaderboard ID | Backing stat | Description |
|---|---|---|
| `atoyr_leaderboard_vanilla` | `atoyr_score_vanilla` | Global vanilla leaderboard. |
| `atoyr_leaderboard_blind` | `atoyr_score_blind` | Global blind leaderboard. |

Tie-breaking must be decided before implementation:

- **Option A**: use a composite score formula (e.g. `score * 1_000_000 + accuracy_pct * 10_000 - end_time_offset`) as the single backing stat and drop the multi-field tie-breaker.
- **Option B**: keep the custom tie-breaker by implementing an AGS Extend Override or by post-processing AGS top-N results in the custom server.

### Analytics events

Fire custom AGS analytics events from the server:

| Event name | When | Payload |
|---|---|---|
| `atoyr.round.started` | `POST /api/v1/game/start` succeeds | `mode`, `itemsUsed`, `roundID`. |
| `atoyr.round.continued` | `POST /api/v1/game/continue` succeeds | `mode`, `roundID`. |
| `atoyr.answer.submitted` | `POST /api/v1/game/submit` returns | `mode`, `roundID`, `correct`, `score`, `attempts`, `remainingSeconds`. |
| `atoyr.round.finished` | timer expires or round ends | `mode`, `roundID`, `score`, `attempts`, `accuracy`, `durationSeconds`. |

Event payloads should be flat JSON key/value pairs.

### IAM roles and permissions

- The server-to-server client needs permissions to write Cloud Save, Statistics, Leaderboards, and Analytics on behalf of players.
- The admin dashboard client needs permissions to read Statistics and Leaderboards for operational dashboards.
- Player IAM clients need permissions to read/write their own Cloud Save records.

On Shared Cloud, permissions follow the `<module>:<group>:<groupId>:<action>` shape. On Private Cloud / BYOC, discover the resource/action strings from the deployed permission catalog instead of assuming Shared Cloud groups.

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
- A minimal round registry in SQLite for deploy recovery and admin counting.
- Request-level metrics, health endpoint, CORS, Sentry, and graceful shutdown.

## Risks and open questions

- **Anonymous vs. authenticated**: moving from a fully anonymous cookie to AGS IAM changes the landing-screen UX and may require a login or silent headless flow.
- **Leaderboard ranking model**: AGS Leaderboards ranks players by statistic value, not individual game sessions. The product must accept one entry per player per mode, or keep a custom session-based leaderboard instead.
- **Leaderboard tie-breaking**: AGS native leaderboards rank by a single statistic. Preserving the current multi-field tie-breaker is the biggest functional gap and likely requires Extend or a composite score.
- **Per-round state**: AGS Session Management is designed for multiplayer match wrappers, not single-player puzzle rounds. Storing round state in Cloud Save is possible but adds latency; keeping it in the custom server is simpler.
- **Operational dashboard**: AGS does not replace the existing request-latency and error-rate dashboard. Decide whether to keep that dashboard or migrate it to a separate observability stack.
- **Cloud Save write failures**: async Cloud Save writes need a retry or dead-letter path so a deploy does not lose the last answer before recovery.
