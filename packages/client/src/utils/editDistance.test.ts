import { describe, expect, it } from 'vitest';
import { editDistance } from './editDistance';

describe('editDistance', () => {
  it('returns 0 for identical strings', () => {
    expect(editDistance('hello', 'hello')).toBe(0);
    expect(editDistance('', '')).toBe(0);
  });

  it('returns correct distance for different strings', () => {
    expect(editDistance('cat', 'cat')).toBe(0);
    expect(editDistance('cat', 'bat')).toBe(1);
    expect(editDistance('cat', 'at')).toBe(1);
    expect(editDistance('cat', 'cart')).toBe(1);
  });

  it('handles empty strings', () => {
    expect(editDistance('', 'abc')).toBe(3);
    expect(editDistance('abc', '')).toBe(3);
  });

  it('returns length of string when other is empty', () => {
    expect(editDistance('hello', '')).toBe(5);
    expect(editDistance('', 'world')).toBe(5);
  });

  it('calculates correct distance for single character differences', () => {
    expect(editDistance('a', 'b')).toBe(1);
    expect(editDistance('ab', 'ac')).toBe(1);
  });

  it('handles scrambled 5-letter words', () => {
    expect(editDistance('hello', 'olleh')).toBe(4);
    expect(editDistance('about', 'tuboa')).toBe(4);
  });
});
