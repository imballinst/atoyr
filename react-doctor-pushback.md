# React Doctor Pushback

Pushback against the recent "doctor" refactor. Agreed on most points; flagged where I still push back.

## 1. Colocated global functions at file bottom — Agreed

`speakLetters` and `getClassNames()` are module-pure and hoisted; moving them to the bottom of `GameScreen.tsx` is fine and reduces noise at the top of the component. Add an AGENTS.md note:

> Module-level helper functions that do not close over component state may be placed at the bottom of the file (function declarations are hoisted). Prefer this over nesting them inside the component.

No counter-argument.

## 2. Removing `marked` — Agreed, dumb

`about.tsx` was converted to hand-written JSX with hardcoded Tailwind classes. This defeats the point of authoring content in Markdown. The right fix is:

- Keep `marked`.
- Keep the custom renderer for Tailwind class hooks.
- Inject via `dangerouslySetInnerHTML` — the markdown source is a static string in the repo, same trusted-input category as GTM snippets, so no sanitization is needed.

Removing the library to avoid `dangerouslySetInnerHTML` is a category error: the threat is untrusted input, not the API itself. The about page source is a static string in the repo, not user- or db-derived content. Restore `marked`.

## 3. `dangerouslySetInnerHTML` for GTM — Agreed, dumb

`root.tsx` GTM snippets were rewritten as literal `<script>{`...`}</script>` and a sandboxed `<iframe>`. Concerns:

- `<script>{templateString}</script>` does **not** execute in React 19 — React renders script children as text content, so the GTM bootstrap never runs. This is a regression in production analytics. The original `dangerouslySetInnerHTML` was correct for a hardcoded, trusted, one-time snippet.
- The `<noscript>` iframe is fine either way; sandboxing is harmless but pointless.

Rule of thumb to codify: `dangerouslySetInnerHTML` is acceptable for **trusted, static, build-time-known** strings (GTM, structured data). It is not acceptable for user/db-derived content. Restore the original GTM implementation.

## 4. Lifting `[answer, setAnswer]` to `useGame` — Agreed

`answer`/`setAnswer` are only consumed by `GameScreen`. Lifting them to `useGame` (and threading through `home.tsx` props) adds surface area with no consumer. Keep the state local to `GameScreen`. The refactor was over-eager "single source of truth" applied where there is only one truth.

## 5. The ref-sync `useEffect` chain — Agreed, this is the worst part

```tsx
const onSubmitRef = useRef(onSubmit);
const tokenRef = useRef(token);
const answerRef = useRef(answer);
useEffect(() => { onSubmitRef.current = onSubmit; });
useEffect(() => { tokenRef.current = token; });
useEffect(() => { answerRef.current = answer; });
```

This is omega slop. Three effects with no deps run every render to keep refs in sync "just in case" the parent passes new identities. We own `useGame`:

- `onSubmit` is `useCallback`-memoized with `[state.phase]` — stable across renders within a phase.
- `token` changes only when the server sends a new word — we can update the ref at that point, or just read from a ref the hook already maintains.
- `answer` is local state we own — we should update `answerRef` **at the mutation site** (`handleLetterClick`, `handleBackspace`), not via an effect.

The pattern "use state to render, use ref to resolve submission" is correct; the implementation is wrong. Refs should be mutated imperatively inside the event handlers that change the corresponding state. Delete all three sync effects.

## 6. Duplicated logic inside the `keydown` effect — Agreed

`submitIfComplete` and `showFeedback` are redefined inside the `useEffect` closure, duplicating the outer `submitIfComplete`/`showFeedback`. Since the handlers already read from refs, we can define `handleKeyDown` once at component scope (memoized or not — we own it, a plain function is fine) and reference `handleLetterClick`/`submitIfComplete` directly. The effect becomes:

```tsx
useEffect(() => {
  function onKeyDown(e: KeyboardEvent) {
    if (e.key === 'Backspace') return handleBackspace();
    const lowerCased = e.key.toLowerCase();
    if (/^[a-z]$/.test(lowerCased)) handleLetterClick(lowerCased);
  }
  window.addEventListener('keydown', onKeyDown);
  return () => window.removeEventListener('keydown', onKeyDown);
}, [handleBackspace, handleLetterClick]);
```

Or, if we want zero deps, register once and read everything through refs (no re-bind). Both are fine; duplicating the logic is not.

## 7. `key={label}` instead of `key={i}` — Agreed

The letter slots are **positional by definition** — slot `i` displays `scrambled[i]`, nothing else. The identity *is* the position; there's no identity independent of the index to preserve. So index keys aren't a downgrade here; they're more honest about what the key represents.

The anti-pattern bites only when both hold:
1. Items can be **reordered/inserted/removed mid-render**, AND
2. Items hold **state worth preserving** across the shuffle (uncontrolled input value, focus, animation progress, component instance state).

Neither holds: the list is fixed-length, never reordered, and the `<div>`s are stateless. There's nothing to misattribute.

Sharper rule: **index keys are correct when position is the identity; semantic keys are correct when the item has an identity independent of its position.** These slots belong to the first category. Using `ORDINAL_LABELS[i]` as the key is fine too — it's stable and reads okay — but it's not doing meaningful work, and `key={i}` would be equally correct. Either is fine; don't rewrite for this reason alone.

---

## Summary of action items

1. Restore `marked` + add `DOMPurify` sanitization to `about.tsx`.
2. Restore `dangerouslySetInnerHTML` GTM snippets in `root.tsx` (the `<script>{`...`}</script>` rewrite does not execute).
3. Revert `answer`/`setAnswer` lift — keep local to `GameScreen`.
4. Delete the three ref-sync `useEffect`s; mutate refs inside event handlers.
5. Deduplicate `submitIfComplete`/`showFeedback` in the `keydown` effect; reuse the component-scope handlers.
6. Add an AGENTS.md note permitting bottom-of-file colocated pure helpers.