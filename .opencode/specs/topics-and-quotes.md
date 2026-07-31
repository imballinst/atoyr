# Topics & Variable-Length Content

> Status: ready to implement

## Goal

Introduce a `topic` axis as a content-category selector and generalize the game away from a hardcoded 5-letter scramble so a new topic — Indonesian politician quotes — can be played as fill-in-the-blank. Topic also becomes a leaderboard partition key, superseding the roadmap's earlier "topic does not partition" stance. The shadcn UI refactor is explicitly **out of scope** for this epic (separate epic).

## Axis Model Update

- `topic` baseline sentinel renames from `common` to **`english-words`**. Existing sessions migrate to `english-words` via the column default.
- `topic` is a single-choice category selector (not stackable), same as today.
- `topic` **partitions leaderboards** — sessions are only compared against sessions of the same topic. This reopens and reverses `roadmap.md`'s "Topic as partition key: no" decision. `roadmap.md` is updated accordingly.
- The roadmap's K-threshold fallback rule (combo pool sparse → degrade) is **not** adopted here; topic boards are hard-partitioned. Revisit if a topic pool is too sparse.

## Topics

- **`english-words`** — the existing 5-letter English scramble content. Mechanic: rearrange the scrambled letters into the answer; definition shown as a hint.
- **`indonesian-politician-quotes`** — fill-in-the-blank quotes attributed to Indonesian politicians. The answer is a single word missing from the quote; the definition is the quote with a placeholder where the blank goes.

The two topics share the same `{ word, definition }` data shape and the same scramble/answer mechanic — they differ only in definition phrasing and answer length. Slot count per turn is the **answer length** (generalized, no longer fixed at 5).

### Definition contract

- The data file's `definition` may contain a literal `<template>` marker marking the blank position. For `english-words` there is no marker; the definition is a plain hint.
- For `indonesian-politician-quotes`, every definition contains exactly one `<template>` marker (e.g. `{ "word": "Kau", "definition": "<template> yang gelap!" }`).
- The server sends `definition` verbatim (including the marker). The client owns rendering:

  - **Visual**: replace `<template>` with `n` underscores where `n = scrambled.length` (the scramble preserves length, so this equals answer length). Render the surrounding definition text around the underscores.
  - **Speech (autoVoice)**: replace `<template>` with a literal `...` (three dots) when building the utterance. The blank's true length is announced to assistive tech via the DOM (an `aria-label` on the blank region stating the number of letters to fill), not via the spoken string.
  - Blind mode interaction remains: the server still strips `definition` to `""` for `mode=blind`; the template logic only runs for `mode=vanilla`.

## Data Layout

- Rename `packages/server/data/` → `packages/server/topics/`.
- Rename `words.json` → `english-words.json` (now `topics/english-words.json`).
- Add `topics/indonesian-politician-quotes.json` (same `{ word, definition }` array shape, definitions contain `<template>`).
- Flat layout — one file per topic, named `<topic>.json`. No per-topic subdirectories or metadata files in this iteration.
- `WordService` loads the whole `topics/` directory at startup and partitions entries by topic (derived from filename sans `.json`). The `WORDS_PATH` env var is replaced with `TOPICS_DIR`. Build/deploy env (Dockerfile, Makefile, deploy workflow) is updated.
- A new sibling skill (alongside `convert-words`) ingests the Indonesian quotes corpus into `indonesian-politician-quotes.json`. The existing `convert-words` skill stays scoped to the 5-letter English format.

## API & Server

- `SessionTopic` enum added: `english-words | indonesian-politician-quotes`. The OpenAPI `StartGameRequest` and `StartGameResponse` gain a `topic: SessionTopic` field. Regenerate `gen.go` and `gen.ts`.
- `POST /api/v1/game/start` validates `req.Topic.Valid()` and forwards it to `SessionService.Create`. `Create` gains a `topic string` param and writes the new column.
- `GameService.StartGame` picks the next word via `wordService.GetRandomWord(session.UsedWords, session.Topic)` — the topic argument filters the in-memory pool.
- `ScrambleWord` and the route's ad-hoc scramble continue to scramble whatever `word` was selected; no length assumption changes server-side.
- `sessionWithVisibleDefinition` and the `SubmitAnswer` definition branch are unchanged — topic is orthogonal to the definition-stripping branch, which stays gated on `mode=blind` only.
- Dead-code cleanup folded in: the `word_definitions` column usage is removed (it was never read or written), and the redundant `CurrentWordToken` assignments in `game.service.go` are deleted.

## Persistence

- Migration `005_session_topic.up.sql`: `ALTER TABLE session_entities ADD COLUMN topic TEXT NOT NULL DEFAULT 'english-words'`. Replace `idx_mode_phase_score_accuracy` with `idx_topic_mode_phase_score_accuracy` (composite on `topic, mode, phase, score, accuracy`) to serve the new partitioned queries. `.down.sql` reverts both.
- `SessionDomain`, the GORM model, and `conversion.go` carry a `Topic` field.

## Leaderboard

- Boards hard-partition by `topic` (in addition to `mode`; `time_constraint` is not yet implemented so it is not part of the key yet).
- `GetLeaderboard(mode, topic, limit, offset)`, `GetTotalEntries(mode, topic)`, and `GetPercentile(sessionID, mode, topic, score)` all gain a `topic` filter in their `WHERE` clauses.
- The leaderboard API response surface gains an optional per-entry `topic` field so a unified board view could badge entries; the client uses it to render a topic chip. (Default client behavior: filter by the active topic rather than show a mixed board.)
- `GetPercentile` tiebreaker ladder is unchanged; only the eligible-pool filter grows a `topic` predicate.

## Client

- `lib/settings.ts`: add `topic: z.enum([...])` to the schema and bump the localStorage key `atoyr:settings:v1` → `:v2` with a migration (existing v1 settings get `topic: 'english-words'`).
- `SettingsModal.tsx`: add a topic selector. Settings now drives `{ autoVoice, mode, topic }`; `apiStartGame` already spreads settings into the start request, so no change there once the OpenAPI request type carries `topic`.
- `GameScreen.tsx` generalization:
  - `ORDINAL_LABELS` is replaced by a length-derived array; the scrambled-word row and the answer-slot row both render `scrambled.length` cells.
  - `submitIfComplete` checks `nextAnswer.trim().length === scrambled.length` (not `5`); `handleLetterClick`'s `>= 5` cap becomes `>= scrambled.length`.
  - Definition rendering replaces `<template>` with `n` underscores (`n = scrambled.length`), keeping surrounding text. An `aria-label` on the blank region discloses the blank length for screen readers (e.g. `blank, 3 letters`).
  - `speakLetters` builds the definition utterance with `<template>` replaced by `...`.
- A new `TopicBanner` component (mirroring `ModeBanner`) renders the active topic during gameplay, rendered at the route level alongside `ModeBanner`.
- `hooks.ts`: `useLeaderboard(mode, topic)` and `useLeaderboardPercentile(mode, topic)` pass `topic` as a query param. `Leaderboard.tsx` exposes a topic selector when `topicProps` is omitted from the call site (mirrors the existing mode-selector pattern).
- `ResultsScreen.tsx` invokes the leaderboard hooks with `settings.topic`.

## Spec / Docs Updates

- Update `.opencode/specs/roadmap.md`: rename baseline `common` → `english-words`; flip "Topic as partition key" to yes; note topic boards are hard-partitioned (no K fallback in this iteration).
- Update root `AGENTS.md` "Implemented Architecture Decisions" with a Topics entry once shipped (and remove this spec file after the user confirms).

## Tests

- Server unit: starting a game with an unknown topic returns a validation error; a known topic stores `topic` on the session.
- Server unit: `GetRandomWord` only returns words from the requested topic's pool and excludes `usedWords`.
- Server unit: `GetLeaderboard`/`GetPercentile`/`GetTotalEntries` filter by `topic`; cross-topic sessions do not contaminate a board.
- Server unit: a `indonesian-politician-quotes` definition is returned verbatim with the `<template>` marker (no server-side rendering).
- Client unit (settings): migrating a v1 settings object yields `topic: 'english-words'`.
- Client unit (GameScreen): variable-length words render `scrambled.length` slots and auto-submit at that length; `<template>` is rendered as `n` underscores with a correct `aria-label`; TTS form uses `...`.
- Client unit (TopicBanner): renders the active topic.
- Client unit (Leaderboard): passes `topic` to the hooks.
- Route-level (home): selecting a topic transitions the landing screen to gameplay with the topic banner visible.

## Out of Scope

- shadcn UI refactor (separate epic).
- `time_constraint` axis (still `survival-mode.md` backlog).
- Per-topic subdirectories / metadata files, and any topic-level scoring weights.
- K-threshold percentile fallback (hard partition only this iteration).
- A mixed-topic unified board view (the API exposes `topic` per entry but the client filters).