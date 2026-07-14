# Analytics Iteration: Faster Replay & Cleaner Landing

## Goal

Increase the replay rate by removing friction from the "Play again" flow, and improve first-time conversion by making the landing screen more compact while keeping the rules discoverable.

## Scope

Three UI/UX changes in the client only:

1. **Instant replay**: clicking "Play again" immediately starts a new game.
2. **How to play modal**: move the rules list from the landing screen into a modal.
3. **Settings modal**: move the auto-voice option off the landing screen into a modal that will host game settings (e.g. mode selection in the future).

## 1. Instant Replay

### Current behavior

The results screen shows a "Play again" button. When clicked, the app returns to the idle/landing screen. The user must then click "Start Game" again to play.

### Desired behavior

- Clicking "Play again" on the results screen starts a new game immediately, same flow as "Start Game".
- A "Back to home" button is added to the results screen, allowing the player to return to the landing screen without replaying.
- The auto-voice preference from the finished session is preserved for the new game.
- The existing `data-ga-label="ga-play-again-button"` tracking attribute is kept so the analytics funnel remains valid.
- The "Back to home" button carries `data-ga-label="ga-back-to-home-button"`.

### Requirements

- The `ResultsScreen` "Play again" button must trigger a new game session rather than returning to idle.
- The `ResultsScreen` must include a "Back to home" button that transitions the app back to the `idle` state (landing screen).
- "Play again" remains the primary call to action; "Back to home" is secondary visually.
- If the new game request fails, the app should stay on the results screen and show a clear error state.
- The home route orchestration must handle the transition from `finished` back to `playing` without an intermediate `idle` state.
- Server calls must reuse the existing start-game API; no new backend endpoints are required.

### Tests

- Unit test for `ResultsScreen`: clicking "Play again" calls the provided replay callback.
- Unit test for `ResultsScreen`: clicking "Back to home" calls the provided home callback.
- Unit test for `useGame` / home route: invoking the replay callback transitions from `finished` to `playing` with a new word and resets the score/timer.
- Unit test for `useGame` / home route: invoking the home callback transitions from `finished` to `idle`.
- Unit test: auto-voice preference from the finished session is passed to the new game.
- Unit test: failed replay keeps the app on the results screen.

## 2. How to Play Modal

### Current behavior

The landing screen (`StartScreen`) displays the full rules inline:

- "You have 30 seconds"
- "Wrong answers cost 1 second"
- "Type using external keyboard or on-screen buttons"

This pushes the "Start Game" button and auto-voice option below the fold on small screens.

### Desired behavior

- The landing screen shows only the title, a one-line pitch, the "Start Game" button, and the two modal triggers ("How to play" and "Settings"). The "Settings" trigger also appears on the results screen next to "Play again".
- A clearly visible "How to play" trigger sits near the "Start Game" button.
- Clicking the trigger opens a modal with the full rules list.
- The modal can be closed with an explicit close control and by clicking outside or pressing Escape.

### Requirements

- Remove the inline rules list from `StartScreen`.
- Add a "How to play" button/link on the landing screen.
- Implement a modal component for the rules. The project already uses `radix-ui`, so use the same primitive family already used elsewhere.
- The modal content must include the existing rules.
- The modal must be accessible: focus trap, close on Escape, close on overlay click, and a visible close button.
- The "Start Game" button remains the primary call to action and must not be visually competing with the modal trigger.

### Tests

- Unit test for `StartScreen`: the rules list is not rendered inline by default.
- Unit test: clicking "How to play" opens the modal and shows the rules.
- Unit test: clicking the close button, overlay, or pressing Escape closes the modal.
- Unit test: the "Start Game" button still starts the game when the modal trigger is present.

## 3. Settings Modal

### Current behavior

The `StartScreen` renders the auto-voice checkbox inline, alongside a footnote explaining the trade-off (5 extra seconds per word, scrambled letters shown only for screen readers). The preference is read from and written to `localStorage` and passed to `onStart`.

### Desired behavior

- The landing screen no longer shows the auto-voice checkbox or its footnote inline.
- A clearly visible "Settings" trigger sits near the "Start Game" button, labeled with text (not just an icon) so the auto-voice option stays discoverable.
- The same "Settings" trigger is also present on the results screen next to "Play again", so players can adjust their settings before replaying.
- Clicking the trigger opens a modal containing the auto-voice toggle and its footnote. The modal is structured to host additional settings later (e.g. a mode selector: "Vanilla" vs "Survival").
- The auto-voice preference continues to persist via `localStorage` so the instant-replay flow can read it without the modal being open.

### Requirements

- Remove the inline auto-voice checkbox and footnote from `StartScreen`.
- Add a "Settings" trigger on the landing screen (next to "Start Game") and on the results screen (next to "Play again"). The settings modal component is shared between both screens.
- Implement a settings modal using the same `radix-ui` primitive family already used elsewhere in the project.
- The modal must be accessible: focus trap, close on Escape, close on overlay click, and a visible close button.
- The auto-voice toggle still reads from and writes to `localStorage`, and the values passed to `onStart` and `onPlayAgain` reflect the current persisted/toggled value.
- The "Start Game" button and "Play again" button remain the primary calls to action and must not visually compete with the "Settings" trigger.

### Tests

- Unit test for `StartScreen`: the auto-voice checkbox and footnote are not rendered inline by default.
- Unit test: clicking "Settings" opens the modal and shows the auto-voice toggle.
- Unit test: toggling auto-voice in the modal persists to `localStorage` and is reflected in the value passed to `onStart`.
- Unit test: closing the modal via close button, overlay, or Escape dismisses it.
- Unit test: the "Start Game" button still starts the game when the "Settings" trigger is present.
- Unit test for `ResultsScreen`: the "Settings" trigger is present next to "Play again".
- Unit test: opening the settings modal from the results screen, toggling auto-voice, and replaying uses the updated persisted value.

## Analytics

New tracking labels are added for the new interactive elements. The existing labels are kept:

- `ga-start-game-button` on the landing screen start button.
- `ga-play-again-button` on the results screen replay button.
- `ga-settings-button` on the "Settings" trigger (both landing and results screen instances).
- `ga-back-to-home-button` on the "Back to home" button on the results screen.

After deployment, verify in GA4 that the funnel `ga-start-game-button` → `ga-play-again-button` improves or stays stable, and that `ga-back-to-home-button` usage is low relative to `ga-play-again-button` (indicating the instant-replay flow is preferred).

## Out of scope

- Changes to the GTM setup.
- New server endpoints or database migrations.
- Changes to scoring, timing, or word selection logic.
- The mode selector UI (Vanilla/Survival) inside the Settings modal; only the modal shell and the auto-voice toggle are in scope.
- The "Survival mode" feature; that is tracked separately in `survival-mode.md`.
