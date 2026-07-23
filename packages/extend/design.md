# Extend design

This document describes the AccelByte Extend patterns needed to work around the gaps in the AGS migration analysis.

## Patterns

| Gap                                    | Pattern                           | Extend app                     | Why                                                                                                                                                   |
| -------------------------------------- | --------------------------------- | ------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| Leaderboard entry metadata             | Service Extension                 | `leaderboard`                  | AGS Leaderboard only stores `(UserID, Rank, Point)`. The Service Extension stores the extra metadata and returns an enriched leaderboard.             |
| Leaderboard total entries / percentile | Service Extension + Event Handler | `leaderboard` + `eventhandler` | AGS has no reliable total count. The Service Extension maintains exact counts and percentile; the Event Handler updates it on every finished session. |
| Session crash recovery                 | Service Extension + Event Handler | `leaderboard` + `eventhandler` | AGS Cloud Save cannot list all keys across users. The Service Extension keeps a small active-session registry.                                        |
| Percentile tiebreak precision          | Service Extension + Event Handler | `leaderboard` + `eventhandler` | AGS ranks by a single numeric point. The Service Extension preserves exact tiebreak columns (score, accuracy, finished-at).                           |

## Architecture

```
┌─────────────┐         AGS API / Cloud Save
│  Atoyr      │◄──────────────────────────────
│  server     │
└──────┬──────┘
       │ Publishes AGS events
       ▼
┌─────────────┐         ┌─────────────┐
│  eventhandler│────────►│  leaderboard │
│  (Extend)    │ updates │  (Extend)    │
└─────────────┘         └──────┬──────┘
                               │
                               ▼
                       ┌─────────────┐
                       │  Atoyr      │
                       │  server     │
                       └─────────────┘
```

The Atoyr server publishes AGS events during the game lifecycle. The `eventhandler` Extend app consumes those events and updates the `leaderboard` Service Extension store. The Atoyr server then queries the Service Extension for:

- `GET /leaderboard?mode={mode}&limit={limit}&offset={offset}` — enriched leaderboard entries.
- `GET /leaderboard/percentile?mode={mode}&userId={userId}` — exact percentile.
- `GET /active-sessions` — registry of currently playing sessions for crash recovery.

## Responsibilities

### `leaderboard` Service Extension

- Store the last finished session metadata per user and mode.
- Maintain exact total entry counts per mode.
- Compute percentile using the same tiebreak semantics as the current SQLite implementation:
  - Higher score first.
  - Higher accuracy second.
  - Earlier `finished_at` third.
- Maintain a registry of active session IDs.
- Expose the three HTTP endpoints above.

### `eventhandler` Event Handler

- Parse AGS events published by the Atoyr server.
- Call the `leaderboard` store to:
  - Add a session ID when a session starts.
  - Remove a session ID and add leaderboard metadata when a session finishes.

## Expected AGS events

The Atoyr server should publish the following events so the Event Handler can keep the Service Extension store in sync.

| Event name               | Payload                                                                     | Handler action                                                                                                                              |
| ------------------------ | --------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `atoyr_session_started` | `{session_id, user_id, mode, started_at}` | Add `session_id` to the active-session registry. |
| `atoyr_session_finished` | `{session_id, user_id, mode, score, total_attempts, accuracy, finished_at}` | Remove `session_id` from the active-session registry; upsert the leaderboard entry for `(user_id, mode)`. |
| `atoyr_round_finished` | `{session_id, user_id, mode, score, total_attempts, accuracy, finished_at}` | Same as `atoyr_session_finished` for round-level updates; the Service Extension treats a finished session as the final authoritative entry. |

## Notes

- The skeletons use an in-memory store so tests can run without MongoDB.
- In production, the `leaderboard` Service Extension would replace the in-memory store with a persistent database (e.g., MongoDB via the AccelByte `nosql-go` patch).
- The Service Extension does not depend on `packages/server`. The Atoyr server communicates with it over HTTP and through AGS events.
