import { test, expect } from 'vitest';
import { resolveQTE } from './index';

test('attacker misses when null', () => {
  expect(resolveQTE(null, 0.5)).toEqual({ kind: 'miss' });
});

test('defender misses -> attacker critical', () => {
  expect(resolveQTE(0.5, null)).toEqual({ kind: 'critical', damage: 20 });
});

test('attacker way faster -> critical', () => {
  const r = resolveQTE(0.2, 0.7);
  expect(r.kind).toBe('critical');
});

test('attacker slightly faster -> hit', () => {
  const r = resolveQTE(0.4, 0.5);
  expect(r.kind).toBe('hit');
});

test('attacker slower -> block', () => {
  const r = resolveQTE(0.7, 0.6);
  expect(r.kind).toBe('block');
});

test('text-based normal hit (attacker accurate & faster)', () => {
  const expected = 'apple';
  // attacker faster and more accurate
  const r = resolveQTE(0.4, 0.6, 'apple', 'appl', expected);
  expect(r.kind).toBe('hit');
  expect(typeof (r as any).damage).toBe('number');
  expect((r as any).damage).toBeGreaterThan(0);
});

test('text-based parry/dodge (defender faster and more accurate)', () => {
  const expected = 'apple';
  // attacker slower and inaccurate, defender fast and accurate -> parry
  const r = resolveQTE(0.7, 0.3, 'x', 'apple', expected);
  expect(r.kind).toBe('parry');
});

test('text-based critical when defender misses (null text)', () => {
  const expected = 'banana';
  const r = resolveQTE(0.5, null, 'banana', null, expected);
  expect(r.kind).toBe('critical');
  expect((r as any).damage).toBeGreaterThan(10);
});
