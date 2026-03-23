import fs from 'fs';
import path from 'path';

const PATH_TO_WORDS_JSON = path.join(process.cwd(), '../server/data/words.json');

let wordsJSON = JSON.parse(fs.readFileSync(PATH_TO_WORDS_JSON, 'utf-8'));
wordsJSON = wordsJSON.filter((entry) => {
  return entry.word.length === 5;
});

fs.writeFileSync(PATH_TO_WORDS_JSON, JSON.stringify(wordsJSON), 'utf-8');
