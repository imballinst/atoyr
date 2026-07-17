# Survival Mode

> Status: backlog / future iteration

## Goal

Add a timer-mechanic variant that rewards speed and endurance. In Survival, every correct answer extends the timer, letting skilled players keep going as long as they stay correct. Survival is orthogonal to `mode` (visibility) and `topic` (content) and may be combined with either.

## Three-Axis Model

The session is no longer described by a single `mode` string. It is composed of three independent axes. Each axis has a baseline sentinel (`vanilla` / `common`) and may grow independently without touching the others.

- **`mode`** — visibility toggle. Values: `vanilla` (default), `blind`. Drives whether the definition is shown. Stackable as a single choice per session (not a set, for now).
- **`time_constraint`** — timer mechanic. Values: `vanilla` (default), `survival`. Drives what the timer does on a correct answer. Survival is a non-baseline mechanic on this axis.
- **`topic`** — content category. Values: `common` (default), and future values (`game`, …). Drives word-pool filtering. Defaulting to `common` means "no topic filter".

This split replaces the current single `mode` column. Migration is done while only one non-baseline value exists per axis (`blind` on `mode`, none yet on `time_constraint`/`topic`), i.e. at the cheap moment. Existing sessions: `mode=vanilla` → `(mode=vanilla, time_constraint=vanilla, topic=common)`; `mode=blind` → `(mode=blind, time_constraint=vanilla, topic=common)`.

### Adding a new value later

- New visibility toggle → new `mode` value; branch in the definition-stripping path only. Other axes untouched.
- New timer mechanic → new `time_constraint` value; branch in the answer/timer path only.
- New content category → new `topic` value; filter the word pool. No other axis touched.

No axis needs to enumerate the others. The combinatorial "every new mode vs every existing one" problem is gone because each axis owns its own branch.

## Gameplay (Survival)

- Starts with the same base timer as `time_constraint=vanilla` (30 seconds).
- Every correct answer adds **+1 second** to remaining time.
- Every wrong answer still costs **1 second**.
- Auto-voice still grants its existing bonus (+5 seconds per correct word) on top of the Survival extension.
- The game ends when the timer reaches zero or all words are exhausted.
- Final score is the number of correct answers before the timer runs out.

## API & Server

- `POST /api/v1/game/start` accepts `mode`, `time_constraint`, and `topic` fields. Each defaults to its baseline sentinel (`vanilla` / `vanilla` / `common`) when omitted, for backward compatibility.
- The session stores all three values in separate columns.
- On a correct answer in a `time_constraint=survival` session, extend the session timer by 1 second via the existing duration manager and persist the new `EndsAt`. The auto-voice bonus (+5s) is added on top, same as today.
- `sessionWithVisibleDefinition` continues to clear the definition iff `mode=blind`, independent of `time_constraint`.
- The submit-answer response continues to return `remainingSeconds`, which reflects the extension.

## Client

- `StartScreen` exposes a `time_constraint` choice (Vanilla / Survival). It already exposes a `mode` choice (Vanilla / Blind). Topic selection is out of scope for this spec.
- `GameScreen` visually indicates when the session is in Survival (e.g. a badge or label). The existing mode banner continues to reflect `mode`.
- The timer behavior is unchanged from the player's perspective; it simply increments on correct answers.
- The results screen shows `time_constraint` so players know which leaderboard they are competing on.

## Leaderboard

Leaderboards are partitioned by `(time_constraint, mode)` — i.e. by real combo — so blind is never compared against vanilla. Topic does not partition leaderboards in this iteration (session topic is stored but not used for filtering).

To handle sparse combo pools at low traffic, `GetPercentile` degrades gracefully:

- Compute the percentile against the combo pool `(time_constraint, mode)`.
- If the combo pool has fewer than **K** finished sessions, fall back to the `(time_constraint)`-only pool, and surface "insufficient blind-survival data; percentile vs all survival players" to the player.
- If the `(time_constraint)`-only pool is also below K, show no percentile rather than a misleading number.
- `GetLeaderboard` and `GetTotalEntries` remain partitioned by `(time_constraint, mode)`; no fallback for the board view itself.

The threshold **K** is a tunable product decision (see Open Questions). It is not a schema fact; document it next to the fallback rule and revisit after launch.

Existing blind board (`mode=blind, time_constraint=vanilla`) is preserved as a contract — it is simply the existing blind board re-keyed under the new schema.

## Analytics

No survival-specific analytics events. The existing `/dashboard` already tracks session data per axis combo; the existing `ga-start-game-button` and `ga-play-again-button` events cover overall funnel tracking.

## Tests

- Server unit test: starting a Survival game stores `time_constraint = survival`.
- Server unit test: a correct answer in Survival extends `EndsAt` and `remainingSeconds` by 1.
- Server unit test: a correct answer in `time_constraint=vanilla` does not extend the timer beyond the existing auto-voice bonus.
- Server unit test: a wrong answer in Survival still decrements the timer by 1.
- Server unit test: Survival + Blind stores both and strips the definition (combination works without per-combo code).
- Server unit test: `GetPercentile` falls back to variant-only pool when the combo pool is below K, and surfaces the degraded state in the response.
- Client unit test: `StartScreen` allows selecting Survival and passes it to `onStart`.
- Client unit test: `GameScreen` displays the Survival indicator.
- Client unit test: `ResultsScreen` displays the `time_constraint`.

## Open Questions

- **Threshold K** for percentile fallback — what value, and is "no percentile" an acceptable UX state below it?
- Should the +1 second reward scale with streak (e.g. +1 for the first 10 correct, then +2)?
- Should Survival have its own color theme or badge to differentiate it visually?
- Should "Play again" in Survival default to Survival, or to the global last-selected `time_constraint`?
- When (if ever) should `topic` start partitioning leaderboards?