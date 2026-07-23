# extend-eventhandler — Implementation Plan

## Target
Event Handler subscribing to custom Atoyr events:
- `atoyr.session.started` — register a new active session
- `atoyr.session.finished` — close session and upsert leaderboard entry
- `atoyr.round.finished` — same as `session.finished` (round-level equivalent)

## Proto provenance
These are **custom telemetry events** published by the Atoyr server, not standard AGS events. The proto is NOT in `github.com/AccelByte/accelbyte-api-proto`.

The event shape uses custom proto messages defined in `pkg/proto/atoyr/events/v1/events.proto`. Generated Go code lives in `pkg/pb/atoyr/events/v1/`.

## Files created
- `pkg/proto/atoyr/events/v1/events.proto` — event message types + AtoyrEventService
- `pkg/proto/leaderboard/v1/leaderboard.proto` — leaderboard client stub (AddSession, RemoveSession, UpsertEntry)
- `pkg/handler/session/handler.go` — JSON-decoding event handler dispatching to leaderboard gRPC client
- `pkg/handler/session/handler_test.go` — tests for session started/finished/round finished/unknown/invalid
- `pkg/leaderboard/client.go` — leaderboard gRPC client interface + implementation

## Files modified
- `main.go` — wired session handler, removed old login/thirdparty handlers, added LEADERBOARD_ADDR env var
- `go.mod` — updated via go mod tidy

## External AGS APIs called
- None directly. The handler calls extend-leaderboard's internal gRPC methods:
  - `AddSession(id, userID, mode, startedAt)`
  - `RemoveSession(id)`
  - `UpsertEntry(mode, entry)`

## Remaining prerequisites
- [x] Confirm event proto approach — custom proto
- [x] Author `pkg/proto/atoyr/events/v1/events.proto`, run `make proto`
- [x] Wire gRPC client to call extend-leaderboard endpoints
- [x] Remove old `packages/extend/eventhandler/` source
- [ ] Register event subscription in Admin Portal for custom event names (`atoyr.session.started`, `atoyr.session.finished`, `atoyr.round.finished`)

### Admin Portal event registration

In the AGS Admin Portal, configure the Event Handler to deliver the following custom telemetry events to the deployed extend-eventhandler gRPC endpoint:

| Event name | Payload shape | Handler action |
|---|---|---|
| `atoyr.session.started` | `{session_id, user_id, mode, started_at}` | AddSession |
| `atoyr.session.finished` | `{session_id, user_id, mode, score, total_attempts, accuracy, finished_at}` | RemoveSession + UpsertEntry |
| `atoyr.round.finished` | `{session_id, user_id, mode, score, total_attempts, accuracy, finished_at}` | RemoveSession + UpsertEntry |

The gRPC service name is `atoyr.events.v1.AtoyrEventService`, method `OnEvent`. The event envelope is `EventEnvelope` with field `event_name` for routing and `payload` for the JSON-encoded event body.

## Out of scope
- Direct database access — all persistence goes through extend-leaderboard's gRPC API
