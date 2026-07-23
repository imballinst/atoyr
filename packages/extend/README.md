# atoyr/extend

AccelByte Extend workarounds for gaps between AGS services and the current Atoyr server implementation.

This package contains skeletons for the Extend apps needed to preserve leaderboard metadata, exact percentile/total counts, and a cross-user active-session registry.

## Structure

- `leaderboard/` — Service Extension that owns custom leaderboard metadata, exact percentile/total counts, and active-session registry.
- `eventhandler/` — Event Handler that consumes AGS events published by the Atoyr server and updates the Service Extension store.
- `design.md` — Patterns, responsibilities, and expected AGS event names.

## Gaps covered

1. **Leaderboard entry metadata** (`TotalAttempts`, `Accuracy`, `Timestamp`) — stored by the Service Extension and returned by `GET /leaderboard`.
2. **Leaderboard total entries and percentile** — exact counts and tiebreak-aware percentile computed in the Service Extension.
3. **Session crash recovery** — active-session registry populated by `atoyr.session.started` / `atoyr.session.finished` events and exposed via `GET /active-sessions`.
4. **Percentile tiebreak precision** — ranking uses the same composite columns (score, accuracy, finished-at) as the current SQLite implementation.

## Development

```sh
make test
make build
```
