Source: WordNet 3.1 (Princeton) — `wordnet/` dir. Download: https://wordnet.princeton.edu/download.

`english-words.json` in `packages/server/topics/` is regenerated from the WordNet drop via `build-wordnet-json.mjs`:

- Candidate pool: 5-letter lowercase lemmas from `index.{noun,verb,adj,adv}`
- Definitions: first gloss from `data.{noun,verb,adj,adv}`, joined on synset offset, cut at first `;`, parentheticals/examples stripped
- Sense selection: WordNet sense order (tagsense-weighted), preferring noun > adj > verb > adv, excluding `noun.person` senses