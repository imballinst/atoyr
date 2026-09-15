import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const MAX_CHARS = 200;
const MAX_ANSWER_CHARS = 9;
const QUOTES_FILE = 'packages/server/topics/indonesian-politician-quotes.json';

try {
  const raw = readFileSync(resolve(QUOTES_FILE), 'utf-8');
  const data = JSON.parse(raw);

  if (!data.words || !Array.isArray(data.words)) {
    throw new Error(`Expected "words" array in ${QUOTES_FILE}`);
  }

  const violations = [];

  for (const entry of data.words) {
    if (entry.word.length > MAX_ANSWER_CHARS) {
      violations.push(
        `  "${entry.word}": answer is ${entry.word.length} chars (max ${MAX_ANSWER_CHARS})`,
      );
    }

    const rawLen = entry.definition.length;
    const resolved = entry.definition.replace(/<template>/g, entry.word);
    const resolvedLen = resolved.length;

    if (rawLen > MAX_CHARS) {
      violations.push(
        `  "${entry.word}": definition is ${rawLen} chars (max ${MAX_CHARS})`,
      );
    }

    if (resolvedLen > MAX_CHARS) {
      violations.push(
        `  "${entry.word}": resolved definition is ${resolvedLen} chars (max ${MAX_CHARS})`,
      );
    }
  }

  if (violations.length > 0) {
    console.error(
      `Validation failed: ${violations.length} quote(s) exceed ${MAX_CHARS} character limit:\n${violations.join('\n')}`,
    );
    process.exit(1);
  }

  console.log(`✓ All ${data.words.length} quotes are within ${MAX_CHARS} character limit and answers within ${MAX_ANSWER_CHARS} character limit`);
} catch (err) {
  if (err.code === 'ENOENT') {
    console.error(`File not found: ${QUOTES_FILE}`);
    process.exit(1);
  }
  throw err;
}
