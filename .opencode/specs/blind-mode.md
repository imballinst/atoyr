# Blind Mode

> Status: backlog / future iteration

## Goal

Add a game mode where the word definition is hidden, increasing the difficulty. Players must reorder the scrambled letters into the correct word without the definition hint.

## Gameplay

- The base timer and scoring rules remain the same as Vanilla mode; Blind mode does not change the scoring formula.
- The definition is not shown to the player at any point during the round.
- Auto-voice remains available, but it only reads the scrambled letters; it does not read a definition.
- Wrong answers still apply the same time penalty.
- The game ends when the timer reaches zero or all words are exhausted.
- The final score is the number of correct answers.
- The word pool is the same across Vanilla and Blind modes.

## Mode Selection

- The landing screen offers a mode choice before starting a game.
- Vanilla remains the default mode and represents the current experience with the definition visible.
- The first version supports only Vanilla and Blind.
- The mode selector should be a separate, prominent control on the landing screen (e.g., tabs or a segmented control near the Start Game button). It is not merged with Settings, which stays focused on player preferences such as auto-voice.
- The selected mode is passed when starting a new game and persisted for the next session (e.g., localStorage).
- "Play again" uses the globally last-selected mode, whether the player started from the landing screen or a previous game.

## API & Server

- `POST /api/v1/game/start` accepts a `mode` field. Valid values are `vanilla` and `blind`. It defaults to `vanilla` when omitted for backward compatibility.
- The session stores the selected mode in a single `mode` text column that defaults to `vanilla`.
- When returning a question for a Blind session, the response omits the definition field.
- Submit-answer logic and scoring are unchanged; the mode only affects the information returned to the player.

## Client

- `StartScreen` shows a mode selector that supports Vanilla and Blind.
- `GameScreen` does not render the definition when the session is in Blind mode.
- `GameScreen` shows a mode banner below the navbar: "🚫 **BLIND MODE** 🚫".
- In Vanilla mode, the banner shows "🍦 **Vanilla mode** 🍦".
- The speak button in Blind mode reads only the scrambled letters, not a definition.
- The results screen shows the mode so players know which leaderboard they are competing on.
- The timer, score, and accuracy remain visible.

## Leaderboard

- Blind mode scores must be tracked separately from Vanilla scores so the two modes do not mix on the same leaderboard.
- The leaderboard page must be updated to cater to multiple modes.
- The page should display a mode selector (or tabs) that lets the player switch between mode-specific leaderboards.
- The default leaderboard view remains Vanilla until the player explicitly selects another mode.
- The percentile comparison on the results screen must compare the player against others in the same mode.

## Analytics

- Add a `ga-blind-mode-started` event when a Blind game begins.
- Add a `ga-blind-mode-finished` event with a `score` parameter when a Blind game ends.
- Keep the existing `ga-start-game-button` and `ga-play-again-button` events for overall funnel tracking.

## Tests

- Server unit test: starting a Blind game stores `mode = blind`.
- Server unit test: starting a game without a mode stores `mode = vanilla`.
- Server unit test: a Blind session question response omits the definition.
- Server unit test: a Vanilla session question response still includes the definition.
- Server unit test: wrong answers in Blind mode still apply the time penalty.
- Client unit test: `StartScreen` allows selecting Blind mode and passes it to `onStart`.
- Client unit test: `GameScreen` does not display the definition in Blind mode.
- Client unit test: `ResultsScreen` displays the mode.
- Client unit test: the leaderboard page shows a mode selector and fetches the leaderboard for the selected mode.
- Integration/client test: a Blind game start returns scrambled letters without the definition.

## Decisions

- The default mode is called **Vanilla** in the UI and stored as `vanilla` in the API/database.
- The in-game mode banner shows:
  - Blind: `🚫 **BLIND MODE** 🚫`
  - Vanilla: `🍦 **Vanilla mode** 🍦`
