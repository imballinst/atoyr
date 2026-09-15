import fs from "fs/promises";
import path from "path";

// Regenerates packages/server/topics/english-words.json from the WordNet 3.1
// drop in scripts/words/wordnet/ (https://wordnet.princeton.edu/download).
//
// Candidate words are 5-letter lowercase lemmas from index.{noun,verb,adj,adv};
// definitions come from the matching data.{noun,verb,adj,adv} gloss, joined by
// synset offset (zero-padded to 8 digits). When a word has several senses, WordNet's
// own tag-ranked sense order is used with a noun > adj > verb > adv preference, and
// noun.person senses are excluded so proper names never appear as answers.

const WORDNET_DIR = path.join(process.cwd(), "scripts/words/wordnet");
const OUTPUT_PATH = path.join(
  process.cwd(),
  "packages/server/topics/english-words.json",
);

const POS_ORDER = ["noun", "verb", "adj", "adv"];

function dataLines(content) {
  return content
    .split("\n")
    .filter((line) => line.length > 0 && !line.startsWith(" "));
}

function parseDataIntoMap(content) {
  const map = new Map();
  for (const raw of dataLines(content)) {
    const parts = raw.split(/\s+/);
    const offset = parts[0];
    if (!/^[0-9]+$/.test(offset)) continue;
    const lexFilenum = Number(parts[1]);
    const glossIdx = raw.indexOf("|");
    const gloss = glossIdx >= 0 ? raw.slice(glossIdx + 1).trim() : "";
    map.set(offset.padStart(8, "0"), { lexFilenum, gloss });
  }
  return map;
}

function sanitizeGloss(gloss) {
  return gloss
    .replace(/"[^"]*"/g, "")
    .replace(/\([^)]*\)/g, "")
    .split(";")[0]
    .trim()
    .replace(/\s+/g, " ");
}

const LEX_FILES = {
  adj: { all: 0, pert: 1, ppl: 3 },
  adv: { all: 2 },
  noun: {
    tops: 3,
    act: 4,
    animal: 5,
    artifact: 6,
    attribute: 7,
    body: 8,
    cognition: 9,
    communication: 10,
    event: 11,
    feeling: 12,
    food: 13,
    group: 14,
    location: 15,
    motive: 16,
    object: 17,
    person: 18,
    phenomenon: 19,
    plant: 20,
    possession: 21,
    process: 22,
    quantity: 23,
    relation: 24,
    shape: 25,
    state: 26,
    substance: 27,
    time: 28,
  },
  verb: {
    body: 29,
    change: 30,
    cognition: 31,
    communication: 32,
    competition: 33,
    consumption: 34,
    contact: 35,
    creation: 36,
    emotion: 37,
    motion: 38,
    perception: 39,
    possession: 40,
    social: 41,
    stative: 42,
    weather: 43,
  },
};

const isPersonSense = (pos, lexFilenum) =>
  pos === "noun" && lexFilenum === LEX_FILES.noun.person;

const POS_PREF = { noun: 0, adj: 1, verb: 2, adv: 3 };

function senseScore(pos, indexRank, tagsense) {
  return POS_PREF[pos] * 1000 - tagsense * 10 + indexRank;
}

(async () => {
  const glossByOffsetByPos = {};
  for (const pos of POS_ORDER) {
    const data = await fs.readFile(path.join(WORDNET_DIR, `data.${pos}`), "utf-8");
    glossByOffsetByPos[pos] = parseDataIntoMap(data);
  }

  const sensesByWord = new Map();
  for (const pos of POS_ORDER) {
    const index = await fs.readFile(path.join(WORDNET_DIR, `index.${pos}`), "utf-8");
    for (const raw of dataLines(index)) {
      // index line layout (WordNet 3.1 files):
      //   A "sense" is one distinct meaning of a word; each sense maps to a synset
      //   (a group of synonymous words sharing that meaning), identified by an
      //   8-digit offset. A word like "light" has 15 senses.
      //   lemma pos synset_cnt pointer_cnt {pointer symbols} tagsense_cnt sense_cnt {synset offsets}
      //   - lemma:           the headword (underscores denote spaces; multiword lemmas are skipped)
      //   - pos:             part of speech: n / v / a / r
      //   - synset_cnt:      how many senses (synsets) this lemma belongs to
      //   - pointer_cnt:     how many pointer symbols follow
      //   - pointer symbols: semantic/lexical relations to other synsets (@ hypernym, ~ hyponym, + derivational, ...)
      //   - tagsense_cnt:    how many of the most common senses carry a semantic-concordance tag count; if 0, no frequency info
      //   - sense_cnt:       total number of senses (must match the offset count)
      //   - synset offsets:  one 8-digit synset id per sense, ordered most-frequent first
      // Example (index.noun):
      //   abuse n 3 3 @ ~ + 3 1 00420921 06728162 00949535
      //   ^^^^^ ^ ^ ^ ^ ^^^ ^ ^ ^^^^^^^^^^^^^^^^^^^^^^^^^^^
      //   lemma | | | |  |  | | sense_cnt=3
      //         | | | |  |  | tagsense_cnt=1
      //         | | | |  |  synset offsets {3}
      //         | | | |  pointer symbols: @ ~ +
      //         | | | pointer_cnt=3
      //         | | synset_cnt=3
      //         | pos='n'
      // The pointer count varies per word, so offsets only start once every
      // pointer symbol of that count has been skipped (here at fields[9+]).
      const fields = raw.split(/\s+/);
      const word = fields[0].replace(/_/g, " ");
      if (!/^[a-z]{5}$/.test(word)) continue;

      const pointerCount = Number(fields[3]);
      if (fields.length <= 4 + pointerCount + 1) continue;

      const tagsenseCount = Number(fields[4 + pointerCount]);
      const senseCount = Number(fields[5 + pointerCount]);
      const offsets = fields
        .slice(6 + pointerCount)
        .filter((f) => /^[0-9]{8}$/.test(f));
      if (offsets.length < senseCount || offsets.length === 0) continue;

      // Glosses live in the data.* file, keyed by synset offset. A word can
      // span the same phrase across POS (e.g. "light" noun/verb/adj), so we
      // collect every sense here and rank them in the pass below.
      const senses = [];
      offsets.forEach((offset, idx) => {
        const record = glossByOffsetByPos[pos].get(offset);
        if (!record) return;

        // idx is the sense's position in the offsets list. WordNet orders senses by
        // semantic-concordance frequency when tags exist (tagsense_cnt > 0),
        // otherwise by lexical-file order — either way rank 0 is the primary
        // sense and later ranks are progressively less common. Only rank 0
        // carries a real tagsense count; the rest default to 0 so they never
        // outrank it.
        const isPerson = isPersonSense(pos, record.lexFilenum);
        senses.push({
          offset,
          pos,
          isPerson,
          indexRank: idx,
          tagsense: idx === 0 ? tagsenseCount : 0,
          lexFilenum: record.lexFilenum,
          gloss: record.gloss,
        });
      });
      if (senses.length === 0) continue;

      const existingSenses = sensesByWord.get(word) ?? [];
      existingSenses.push(...senses);
      sensesByWord.set(word, existingSenses);
    }
  }

  const words = [];
  for (const [word, senses] of sensesByWord) {
    // Pick one gloss per word: POS preference (noun > adj > verb > adv) beats
    // tag rank, so a common noun/adj sense wins over a verb sense even when
    // WordNet ranks the verb first. Person senses are dropped outright.
    const usable = senses
      .filter((s) => !s.isPerson)
      .map((s) => ({
        ...s,
        score: senseScore(s.pos, s.indexRank, s.tagsense),
      }))
      .sort((a, b) => a.score - b.score);
    const best = usable[0];
    if (!best) continue;

    const definition = sanitizeGloss(best.gloss);
    if (definition.length < 5) continue;

    words.push({ word, definition });
  }

  words.sort((a, b) => a.word.localeCompare(b.word));

  await fs.writeFile(
    OUTPUT_PATH,
    `${JSON.stringify({ lang: "en-US", words }, null, 2)}\n`,
  );

  console.log(`wrote ${words.length} words -> ${OUTPUT_PATH}`);
})();