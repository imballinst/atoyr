import { SCRAMBLE_ATTEMPTS } from '../types';
import { editDistance } from './editDistance';

function shuffleLetters(word: string): string {
  const letters = word.split('');
  for (let i = letters.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [letters[i], letters[j]] = [letters[j], letters[i]];
  }
  return letters.join('');
}

export function scrambleWord(word: string): string | null {
  const shuffles: string[] = [];
  for (let i = 0; i < SCRAMBLE_ATTEMPTS; i++) {
    shuffles.push(shuffleLetters(word));
  }

  const validShuffles = shuffles.filter((s) => s !== word);

  if (validShuffles.length === 0) return null;

  return validShuffles.sort((a, b) => editDistance(b, word) - editDistance(a, word))[0];
}
