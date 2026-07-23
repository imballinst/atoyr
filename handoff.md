# AGS integration — handoff

Completed:

- IAM clients created: server-to-server (confidential `e50254a28f0740369981d194137abd14`) and web/public (`cecf5228881640ef88fc9a3fbc079cc3`)
- Server-to-server client permissions added: Cloud Save (Player Records CRU), Statistics (User Values CU), Leaderboard (Config/Data R, User Visibility R)
- AccelByte Go SDK installed and wired into `packages/server/cmd/service/main.go`
- `.env.local`/`.env.local.example` updated with AGS config

## Next steps

- ✅ **Create Cloud Save / Statistics / Leaderboard configurations in AGS** (`/ags manage-resource`) — completed. AGS stat codes and leaderboard codes use hyphens because underscores are rejected by AGS validation:
  - Stats: `atoyr-score-vanilla`, `atoyr-score-blind`, `atoyr-accuracy-vanilla`, `atoyr-accuracy-blind`, `atoyr-attempts-vanilla`, `atoyr-attempts-blind`, `atoyr-composite-vanilla`, `atoyr-composite-blind` (all `setBy: SERVER`, `visibility: SERVERONLY`).
  - Leaderboards: `atoyr-leaderboard-vanilla` (backed by `atoyr-composite-vanilla`, `descending: true`, `allTime: true`), `atoyr-leaderboard-blind` (backed by `atoyr-composite-blind`).
- ✅ **Decide on the leaderboard tie-breaker strategy** — Option A (composite score formula). `atoyr-composite-*` encodes `score * 1_000_000 + accuracy * 10_000 - durationSeconds` so AGS native leaderboards can rank by a single stat. Higher score wins; equal score, higher accuracy wins; equal accuracy, faster finish wins.
- ✅ **Implement the actual migration of `game.service.go`** endpoints to use AGS Cloud Save (round persistence), Statistics (post-game score/attempts/accuracy/composite), and Leaderboards (top-N/percentile queries):
  - `internal/platform/accelbyte/accelbyte.go` now exposes Cloud Save, Statistics, Leaderboard, and IAM headless-account service clients.
  - `internal/services/agssync.service.go` wraps Cloud Save player-record writes/reads, bulk Statistics updates, and Leaderboard top-N/user-rank queries.
  - `internal/services/game.service.go` creates an AGS headless account on first start, writes round state to Cloud Save, and posts final score/attempts/accuracy/composite stats on finish (async where appropriate).
  - `internal/services/leaderboard.service.go` falls back to SQLite when AGS is disabled; when enabled it reads top-N and user rank from AGS.
  - `session_entities` gained a `user_id` column (migration `005_add_user_id`) and `SessionDomain`/`SessionEntity` carry the AGS player ID.
  - Added `internal/services/agssync_test.go` covering the composite-score calculation and disabled-service behavior.

## Open follow-ups

- **Client AGS token flow**: the server currently creates headless accounts internally and stores the AGS `user_id` in the SQLite session. To support cross-device resume or explicit login, the client should receive and send the IAM access token (e.g. in an `Authorization` header or cookie) and the server should validate it via IAM introspection.
- **Admin dashboard AGS auth**: replace Google OAuth with AGS IAM admin roles/permissions for `/admin/*`.
- **AGS Analytics events**: fire `atoyr.round.started`, `atoyr.answer.submitted`, and `atoyr.round.finished` through `gametelemetry`.
- **Health check AGS dependency**: optionally include AGS token/Cloud Save reachability in `/api/v1/health`.
- **Cloud Save write retry**: add a retry/dead-letter path for failed async Cloud Save writes so deploy recovery does not lose the latest answer.
