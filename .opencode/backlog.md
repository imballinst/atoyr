# Backlog

## Global game state with Jotai

### Motivation

`ResultsScreen.test.tsx` passes mocked `onPlayAgain`/`onBackToHome` callbacks and asserts they were called — this tests internal wiring, not user-observable behavior. With global state, the test could instead:

1. Seed atom values (score, phase, currentWord, etc.)
2. Click "Play Again"
3. Assert that phase transitions to `"playing"` (i.e. the results screen is replaced by the game screen)

No callback mocks needed.

### Current architecture

- `useGame` hook in `packages/client/app/api/hooks.ts` owns all game state via `useState<GameState>`
- `Home` route (`packages/client/app/routes/home.tsx`) calls `useGame(shouldFetch)` and spreads the returned `state` + callbacks to child screens as props
- `ResultsScreen` receives `onPlayAgain: () => void` and `onBackToHome: () => void` as props — the only reason these exist is so `Home` can wire them up
- No client-state library is used today

### Proposed changes

1. **Add jotai** as a dependency in `packages/client`
2. **Create `packages/client/app/lib/game-atoms.ts`** with atoms:
   - `gameStateAtom` — writable atom holding `GameState`
   - Derived atoms for individual slices (`gamePhaseAtom`, `scoreAtom`, etc.) if needed
3. **Refactor `useGame`** to read from and write to atoms instead of internal `useState`:
   - `setState(...)` → `useSetAtom(gameStateAtom)`
   - Internal reads change from `state.xxx` to `useAtomValue(gameStateAtom)`
   - SSE subscription, `sessionRef`, `isBeforeUnloadRef`, etc. stay as `useRef` inside the hook
4. **Refactor `Home`** to pass fewer props (or none) — child screens read directly from atoms:
   - `ResultsScreen` reads `gameStateAtom` directly, no `onPlayAgain`/`onBackToHome` props
   - `ResultsScreen` calls `useSetAtom()` to transition phase to `"playing"` (play again) or `"idle"` (back to home)
   - Similarly for `StartScreen` and `GameScreen`
5. **Rewrite `ResultsScreen.test.tsx`** to:
   - Set `gameStateAtom` to a finished state
   - Assert on rendered content (score, last word, buttons present)
   - Click "Play Again" → assert phase atom changed to `"playing"` (or that results screen is gone)

### Tradeoffs

**Pros:**
- Tests become resilient — they verify what the user sees, not internal wiring
- Components are decoupled from `Home`'s prop wiring
- Jotai is lightweight (~3KB) and has no boilerplate (no providers, reducers, actions)

**Cons:**
- Adds a dependency for what is essentially one state object
- Global state makes data flow less explicit than props
- SSE lifecycle management (`useRef`-based unsubscribe) in the hook stays imperative regardless of state approach — atoms don't simplify that part
- The current prop-drilling is shallow (only one level: `Home` → child screens); the wiring cost is negligible
- Two components (`ResultsScreen`, `GameScreen`) currently read `GameState` values — they'd need `useAtomValue` imports everywhere instead of receiving props; this may _reduce_ colocation

### Alternative: keep `useGame`, test via `Home` route

Instead of global state, render the `Home` route in tests and mock the API layer (MSW or similar). This keeps the existing architecture and tests user-observable paths (click "Play Again" → game screen appears). The tradeoff is that tests become integration-level and slower to write/run.

### Decision needed

Which approach to pursue:

- **A) Jotai global state** — more resilient unit tests, new dependency, less explicit data flow.
- **B) Test via `Home` route (MSW)** — keep current architecture, slower but more realistic tests.
- **C) Keep current tests** — the "callback-assertion" pattern works; the awkwardness is tolerable.
