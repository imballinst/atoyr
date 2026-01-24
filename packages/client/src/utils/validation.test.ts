import { describe, expect, it } from 'vitest';
import { validateAnswer } from './validation';

describe('validateAnswer', () => {
  it('validates exact match', () => {
    expect(validateAnswer('hello', 'hello')).toBe(true);
    expect(validateAnswer('world', 'world')).toBe(true);
  });

  it('is case insensitive', () => {
    expect(validateAnswer('HELLO', 'hello')).toBe(true);
    expect(validateAnswer('HeLLo', 'hello')).toBe(true);
    expect(validateAnswer('hello', 'HELLO')).toBe(true);
  });

  it('trims whitespace', () => {
    expect(validateAnswer('  hello  ', 'hello')).toBe(true);
    expect(validateAnswer('hello', '  hello  ')).toBe(true);
    expect(validateAnswer(' HELLO ', 'hello')).toBe(true);
  });

  it('returns false for non-matching answers', () => {
    expect(validateAnswer('hello', 'world')).toBe(false);
    expect(validateAnswer('cat', 'dog')).toBe(false);
  });

  it('returns false for partial matches', () => {
    expect(validateAnswer('hel', 'hello')).toBe(false);
    expect(validateAnswer('hello', 'hell')).toBe(false);
  });

  it('returns false for scrambled answers', () => {
    expect(validateAnswer('olleh', 'hello')).toBe(false);
    expect(validateAnswer('world', 'dlrow')).toBe(false);
  });
});
