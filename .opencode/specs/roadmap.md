# Roadmap

Forward-looking doc for Atoyr. Not a build spec — it tracks the planned expansion axes and the backlog of future values/features per axis so design decisions are made once, not re-litigated.

## Three-Axis Session Model

Every session is described by three independent axes. The baseline sentinel on each is `vanilla` (`common` for topic). New values slot into one axis without touching the others.

| Axis | Column | Drives | Baseline | Current non-baseline values |
|---|---|---|---|---|
| **`mode`** | visibility toggle | whether the definition is shown (definition-stripping branch) | `vanilla` | `blind` |
| **`time_constraint`** | timer mechanic | what the timer does on a correct/wrong answer (answer/timer branch) | `vanilla` | — (survival planned, see `survival-mode.md`) |
| **`topic`** | content category | which words are eligible (word-pool filter) | `common` | — |

### Rules of the model

- Each axis owns exactly one branch site in the server. Adding a value on one axis never requires changes on another.
- No axis enumerates the others. The combinatorial "every new value vs every existing one" problem does not exist under this split.
- Leaderboards partition by `(time_constraint, mode)` with graceful fallback to `time_constraint`-only below threshold K (see `survival-mode.md`). Topic does not partition leaderboards in this iteration.
- A session may combine any value on each axis, e.g. `mode=blind, time_constraint=survival, topic=common`.

## Backlog

### `mode` axis (visibility)

- `blind` — **shipped**. Strips the definition.
- *No other values currently planned.* Candidate ideas (not committed): `no-scramble-preview` (hide the scrambled word order preview), `hardcore` (no retries per word). Any new toggle goes here, not on `time_constraint`.

### `time_constraint` axis (timer mechanic)

- `vanilla` — baseline. Current default behavior.
- `survival` — **planned**, see `survival-mode.md`. +1s per correct, wrong still −1s.
- *No other values currently planned.* Any non-time mechanic (attempts-per-word, word-length gating) does **not** belong here — it would need its own axis. If one ever shows up, re-open the axis model rather than overloading this column.

### `topic` axis (content category)

- `common` — baseline. No filter; uses the default word pool.
- Future values under consideration: `game`, … (to be defined when the word pool grows enough to support topic filtering). Topic values are a category selector, not a stackable toggle — a session picks one topic (or `common`).
- Open: should `topic` ever partition leaderboards? Only revisit if a topic becomes dense enough to support its own boards.

## Non-Axis Backlog

Features outside the three-axis model.

- **Speech / auto-voice** — shipped. Toggleable, persisted in localStorage. Not an axis; a per-session preference.
- **Monitoring dashboard** (`/dashboard`) — shipped. No planned expansion beyond the existing metrics/timeseries.
- **Healthcheck** — shipped.
- *(add future non-axis features here as they are scoped)*

## Decisions to Make Once

These are living decisions, not schema facts. Document the chosen value next to each and revisit only with evidence.

- **K** (percentile-fallback threshold): value TBD after launch traffic observation. Currently unspecified.
- **Leaderboard partition key**: `(time_constraint, mode)` with fallback. Revisit if a topic becomes dense.
- **Topic as partition key**: no, for this iteration.

## Out of Scope for Now

- Stackable `mode` (a session choosing multiple visibility toggles at once). Model today is single-choice per axis. If two visibility toggles ever need to coexist, promote `mode` to a set before adding the second toggle — do not bolt it on.
- More than three axes. If the model stops fitting, re-open the axis design rather than overloading a column.