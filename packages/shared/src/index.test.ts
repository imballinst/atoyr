import { test, expect } from 'vitest';
import { resolveQTE } from './index';

test('attacker misses when null', () => {
  expect(resolveQTE(null, { time: 0.5 })).toEqual({ kind: 'miss' });
});

test('defender misses -> attacker critical', () => {
  expect(resolveQTE({ time: 0.5 }, null)).toEqual({ kind: 'critical', damage: 15 });
});

test('attacker way faster -> critical', () => {
  const r = resolveQTE({ time: 0.2 }, { time: 0.7 });
  expect(r.kind).toBe('critical');
});

test('attacker slightly faster -> hit', () => {
  const r = resolveQTE({ time: 0.4 }, { time: 0.5 });
  expect(r.kind).toBe('hit');
});

test('attacker slower -> block', () => {
  const r = resolveQTE({ time: 0.7 }, { time: 0.6 });
  expect(r.kind).toBe('block');
});

test('text-based normal hit (attacker accurate & faster)', () => {
  const expected = 'apple';
  // attacker faster and more accurate
  const r = resolveQTE({ time: 0.4, text: 'apple' }, { time: 0.6, text: 'appl' }, expected);
  expect(r.kind).toBe('hit');
  expect(typeof (r as any).damage).toBe('number');
  expect((r as any).damage).toBe(11);
});

test('text-based parry/dodge (defender faster and more accurate)', () => {
  const expected = 'apple';
  // attacker slower and inaccurate, defender fast and accurate -> parry
  const r = resolveQTE({ time: 0.7, text: 'x' }, { time: 0.3, text: 'apple' }, expected);
  expect(r.kind).toBe('parry');
});

test('text-based critical when defender misses (null text)', () => {
  const expected = 'banana';
  const r = resolveQTE({ time: 0.5, text: 'banana' }, null, expected);
  expect(r.kind).toBe('critical');
  expect((r as any).damage).toBe(15);
});

test("attacker empty string vs defender correct 'banana' -> defender parry", () => {
  const expected = 'banana';
  const r = resolveQTE({ time: 0.5, text: '' }, { time: 0.4, text: 'banana' }, expected);
  expect(r.kind).toBe('parry');
  expect(typeof (r as any).damage).toBe('number');
  expect((r as any).damage).toBe(7);
});
