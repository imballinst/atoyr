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
