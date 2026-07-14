# Survival Mode

> Status: backlog / future iteration
>
> Tentative name: "Survival". Other candidates: Marathon, Infinite, Zen, Time Attack.

## Goal

Add a second game mode that rewards speed and endurance. In Survival mode, every correct answer extends the timer by a small amount, allowing skilled players to keep playing as long as they can answer correctly.

## Gameplay

- The player starts with the same base timer as Classic mode (30 seconds).
- Every correct answer adds **+1 second** to the remaining time.
- Every wrong answer still costs **1 second**, same as Classic.
- Auto-voice still grants its existing bonus (+5 seconds per correct word) in addition to the Survival extension.
- The game ends when the timer reaches zero or all words are exhausted.
- The final score is the number of correct answers before the timer runs out.

## Mode Selection

- The landing screen offers a mode choice before starting.
- Classic remains the default mode.
- The selected mode is passed when starting a new game.
- The mode selection is preserved for the next session (e.g. localStorage), so returning players start with their last chosen mode.

## API & Server

- `POST /api/v1/game/start` accepts a `mode` field (`classic` or `survival`). Default to `classic` when omitted for backward compatibility.
- The session stores the selected mode.
- When processing a correct answer in a Survival session, extend the session timer by 1 second via the existing duration manager and persist the new `EndsAt`.
- The submit-answer response continues to return `remainingSeconds`, which will reflect the extension.

## Client

- `StartScreen` shows a mode selector (Classic / Survival).
- `GameScreen` visually indicates that the session is in Survival mode (e.g. a badge or label).
- The timer behavior is unchanged from the player's perspective; it simply increments on correct answers.
- The results screen shows the mode so players know which leaderboard they are competing on.

## Leaderboard

- Survival mode scores must be tracked separately from Classic scores so the two modes do not mix on the same leaderboard.
- Options to decide during implementation:
  1. Add a `mode` column to leaderboard entries and filter the leaderboard by mode.
  2. Maintain two separate leaderboard endpoints/tables.
- The percentile comparison on the results screen must compare the player against others in the same mode.

## Analytics

- Add a `ga-survival-mode-started` event (or equivalent label) when a Survival game begins.
- Add a `ga-survival-mode-finished` event with a `score` parameter when a Survival game ends.
- Keep the existing `ga-start-game-button` and `ga-play-again-button` events for overall funnel tracking.

## Tests

- Server unit test: starting a Survival game stores `mode = survival`.
- Server unit test: a correct answer in Survival mode extends `EndsAt` and `remainingSeconds` by 1.
- Server unit test: a correct answer in Classic mode does not extend the timer beyond the existing auto-voice bonus.
- Server unit test: a wrong answer in Survival mode still decrements the timer by 1.
- Client unit test: `StartScreen` allows selecting Survival mode and passes it to `onStart`.
- Client unit test: `GameScreen` displays the Survival mode indicator.
- Client unit test: `ResultsScreen` displays the mode.
- Integration/client test: a Survival game start returns `remainingSeconds` reflecting the base timer.

## Open Questions

- Should the +1 second reward scale with difficulty or streak? For example, +1 for the first 10 correct answers, then +2, etc.
- Should Survival mode have its own color theme or badge to differentiate it visually?
- Should the leaderboard show both modes with a toggle, or should the mode selector on the leaderboard page be independent?
- Should a "Play again" in Survival mode default to Survival or to the global last-selected mode?
