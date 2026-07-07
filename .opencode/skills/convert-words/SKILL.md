---
name: convert-words
description: Convert scripts/words/output.txt into the words.json database with short definitions, filtering out people names
license: MIT
compatibility: opencode
metadata:
  audience: maintainers
  workflow: data
---

## What I do

- Read `scripts/words/output.txt` (text file, one 5-letter word per line)
- Replace existing definitions in `packages/server/data/words.json` so that we use fully the words from latest source
- Generate short definitions for new words (max 5 words OR max 35 characters)
- The struct is an array of `{ word: string, definition: string }`
- Filter out words that are primarily people names
- Write valid JSON to `packages/server/data/words.json`
- Run `oxlint` to verify

## When to use me

Use this when you need to refresh or expand the game's word bank from the raw word list. The source file can be regenerated from a dictionary or corpus and this skill will re-merge definitions without losing curated ones.
