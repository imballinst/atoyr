# Share Settings

> Status: ready for implementation

## Goal

Let a player share their game settings via a URL so a recipient's client adopts them on visit, without disrupting an in-progress game. The recipient's settings update silently; the share URL is cleaned up before the page settles.

## Share affordance

- A "Share settings" button in `SettingsModal` (`packages/client/app/components/SettingsModal.tsx`), visible for both the landing-screen and results-screen instances of the modal.
- On click: build `${location.origin}/?settings=${btoa(JSON.stringify(settings))}` and copy it to the clipboard via `navigator.clipboard.writeText`. Show transient "Copied!" feedback that reverts after ~2s.
- No analytics events. The existing dashboard already covers session tracking; share is not a game lifecycle event.

## URL contract

- The `settings` query param is the base64 of a JSON `LatestSchema` (`{ mode, topic, autoVoice }`). The payload is ASCII-only, so plain `btoa` is sufficient (no `encodeURIComponent` dance needed).
- The share link always targets `/` (root). Applying settings is the responsibility of the `home.tsx` `clientLoader` only.

## Apply + strip (home clientLoader)

`packages/client/app/routes/home.tsx` `clientLoader`:

1. Read the `settings` search param. If absent, behavior is unchanged.
2. If present, compute `gameInProgress = !hasGameEnded()` (`packages/client/app/lib/game.ts:58`). This is the same localStorage-only signal already used to gate `shouldFetch`—no API call. It is a faithful proxy: when `gameEndsAt` is in the future the player is mid-game; once it's past, the server session has ended too, so resume would not override the applied settings.
3. If `!gameInProgress`: `atob` → `JSON.parse` → `V2SettingsSchema.safeParse`. On success, `writeStoredSettings(parsed)` (wholesale replace, not a merge). On any failure (bad base64, bad JSON, schema-violating), ignore silently—no partial write.
4. If `gameInProgress`: ignore the payload (do nothing).
5. Always strip the `settings` param via `window.history.replaceState(null, '', pathname + hash)` when it is present—even when ignored due to an in-progress game—so the URL is clean before the loader returns.
6. Return `{ shouldFetch, settings: readStoredSettings() }` as today. The newly written settings flow into `state.settings` through the existing `useGame` seeding path, consistent with how `updateSettings` already works.

## Replace, not merge

The payload is the full `LatestSchema`. Apply is wholesale replace. When the schema grows later, the share generator emits the full new schema; older payloads missing new required fields fail `safeParse` and are ignored rather than applied partially.

## Tests

- Unit (`lib/settings.ts`): add pure `encodeSettings(s) → string` and `decodeSettings(str) → LatestSchema | null` and test round-trip plus invalid inputs: malformed base64, invalid JSON, schema-violating values.
- Component (`SettingsModal.test.tsx`): clicking "Share settings" calls `navigator.clipboard.writeText` with the expected `${origin}/?settings=...` URL and surfaces the visible "Copied!" state. Use a semantic query for the button; assert the visible feedback.
- Route (`home.test.tsx`, per the assert-transitions/outcomes convention): visiting `/?settings=<valid blob>` with no in-progress game renders the shared `mode`/`topic` indirectly (e.g., the ModeBanner shows the shared combo); with an in-progress game (`gameEndsAt` set in the future) the shared settings do **not** take effect.

## Out of scope

- Server-side share state; authenticated share; per-field selective share UI; share-from-any-route (share link is root-only).