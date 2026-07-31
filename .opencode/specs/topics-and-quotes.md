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

---

## Task Breakdown

### Phase 1 — Data Layer & Persistence

**1.1 Rename data directory and words file**
- Rename `packages/server/data/` → `packages/server/topics/`
- Rename `packages/server/topics/words.json` → `packages/server/topics/english-words.json`
- Update Dockerfile, Makefile, and env configs: replace `WORDS_PATH` with `TOPICS_DIR`

**1.2 Migration 005 — add `topic` column + new index**
- Create `005_session_topic.up.sql`: `ALTER TABLE session_entities ADD COLUMN topic TEXT NOT NULL DEFAULT 'english-words'`; drop `idx_mode_phase_score_accuracy`, create `idx_topic_mode_phase_score_accuracy(topic, mode, phase, score, accuracy)`
- Create `005_session_topic.down.sql`: revert both

**1.3 Add `Topic` to GORM model and domain model**
- `models/models.go`: add `Topic string` field to `SessionEntity`
- `domainmodels/conversion.go`: add `Topic string` to `SessionDomain`; carry through conversion functions

**1.4 Refactor WordService for multi-topic loading**
- `word.service.go`: replace `WORDS_PATH` env with `TOPICS_DIR`; load entire directory at startup; partition entries by filename stem; `GetRandomWord` gains `topic string` param and filters the in-memory pool
- Remove dead code: `word_definitions` column usage

**1.5 Update SessionService.Create to accept `topic`**
- `session.service.go`: add `topic string` param, write to new column on session creation

### Phase 2 — API & Codegen

**2.1 Update OpenAPI spec**
- `api.yaml`: add `SessionTopic` enum (`english-words | indonesian-politician-quotes`); add `topic: SessionTopic` to `StartGameRequest` and `StartGameResponse`; add `topic` query param to leaderboard endpoints

**2.2 Regenerate server and client codegen**
- Run `go tool oapi-codegen` for `gen.go`
- Run `openapi-typescript` for `gen.ts`

### Phase 3 — Server Logic

**3.1 Validate topic on game start**
- `game_routes.go`: validate `req.Topic.Valid()`; forward to `SessionService.Create` and `GameService.StartGame`

**3.2 GameService passes topic to WordService**
- `game.service.go`: pass `session.Topic` to `wordService.GetRandomWord`
- Remove redundant `CurrentWordToken` assignments

**3.3 Update leaderboard service for topic partitioning**
- `leaderboard.service.go`: add `topic` param to `GetLeaderboard`, `GetTotalEntries`, `GetPercentile`; add `AND topic = ?` to WHERE clauses
- `leaderboard_routes.go`: read `topic` from query params, default to `"english-words"`
- `conversion.go`: add `Topic` to `ToApiLeaderboardEntry`

### Phase 4 — Client: Settings & Banners

**4.1 Update settings with topic field**
- `lib/settings.ts`: add `topic: z.enum([...])` to schema; bump localStorage key to `atoyr:settings:v2`; add migration (v1 → v2 injects `topic: 'english-words'`)

**4.2 Add topic selector to SettingsModal**
- `SettingsModal.tsx`: add `<select>` for topic (mirrors existing mode selector)

**4.3 Create TopicBanner component**
- New `components/TopicBanner.tsx`: mirrors `ModeBanner` pattern — renders active topic name

**4.4 Render TopicBanner at route level**
- `routes/home.tsx`: render `TopicBanner` alongside `ModeBanner` during gameplay

### Phase 5 — Client: GameScreen Generalization

**5.1 Replace hardcoded length assumptions**
- `GameScreen.tsx`: remove `ORDINAL_LABELS`; generate ordinal array from `scrambled.length`; `submitIfComplete` checks `=== scrambled.length`; `handleLetterClick` caps at `scrambled.length`; slot rows render `scrambled.length` cells

**5.2 Definition `<template>` replacement (visual)**
- `GameScreen.tsx`: replace `<template>` with `n` underscores; keep surrounding text; add `aria-label="blank, N letters"` on the underscore region

**5.3 Definition `<template>` replacement (speech)**
- `GameScreen.tsx` — `speakLetters`: replace `<template>` with `...` in the definition utterance

### Phase 6 — Client: Leaderboard

**6.1 Add topic param to leaderboard hooks**
- `hooks.ts`: `useLeaderboard` and `useLeaderboardPercentile` gain `topic` param

**6.2 Pass topic from settings to leaderboard**
- `ResultsScreen.tsx`: invoke leaderboard hooks with `settings.topic`
- `Leaderboard.tsx`: expose topic selector; render topic chip per entry

### Phase 7 — Tests

**7.1 Server unit tests**
- Unknown topic returns validation error; known topic stores `topic` on session
- `GetRandomWord` only returns words from requested topic and excludes `usedWords`
- `GetLeaderboard`/`GetPercentile`/`GetTotalEntries` filter by `topic`; cross-topic sessions don't contaminate
- Indonesian-politician-quotes definition is returned verbatim with `<template>` marker

**7.2 Client unit tests**
- Settings: migrating v1 → v2 yields `topic: 'english-words'`
- GameScreen: variable-length words render correct slot count and auto-submit at that length; `<template>` → `n` underscores with correct `aria-label`; TTS uses `...`
- TopicBanner: renders active topic
- Leaderboard: passes `topic` to hooks

**7.3 Route-level test**
- `home.test.tsx`: selecting a topic transitions landing screen to gameplay with topic banner visible

### Phase 8 — Docs & Cleanup

**8.1 Update roadmap.md**
- Rename baseline `common` → `english-words`; flip "Topic as partition key" to yes; note hard-partition (no K fallback)

**8.2 Update AGENTS.md**
- Add Topics entry under "Implemented Architecture Decisions"; remove this spec file (after user confirms)