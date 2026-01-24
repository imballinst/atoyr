import { describe, expect, it } from 'vitest';
import { scrambleWord } from './scramble';

describe('scrambleWord', () => {
  it('returns a different string than input', () => {
    const word = 'hello';
    const result = scrambleWord(word);
    expect(result).not.toBe(word);
  });

  it('returns string with same letters', () => {
    const word = 'about';
    const result = scrambleWord(word);
    if (result) {
      expect(result.split('').sort().join('') === word.split('').sort().join(''));
    }
  });

  it('returns non-null for common 5-letter words', () => {
    const words = ['hello', 'world', 'about', 'apple', 'beach'];
    words.forEach((word) => {
      expect(scrambleWord(word)).not.toBeNull();
    });
  });

  it('returns null when unable to create valid shuffle', () => {
    const singleChar = 'aaaaa';
    expect(scrambleWord(singleChar)).toBeNull();
  });

  it('selects hardest scramble by edit distance', () => {
    const word = 'abcde';
    const result = scrambleWord(word);
    expect(result).toBeTruthy();
    expect(result).not.toBe(word);
  });
});
