import { describe, it, expect, beforeEach } from 'vitest';
import { MatchService } from './match.service';

function makePlayers(ms: MatchService) {
  const a = ms.createPlayer('pA', 'Alice', 1);
  const b = ms.createPlayer('pB', 'Bob', 1);
  return { a, b };
}

describe('MatchService QTE & HP', () => {
  let ms: MatchService;

  beforeEach(() => {
    ms = new MatchService();
  });

  it('both players hp remain numeric after a fast attacker hits', () => {
    const { a, b } = makePlayers(ms);
    const match = ms.matchmake(a.id, b.id);
    // ensure match started and a QTE exists
    const payload = ms.startQTE(match.id, a.id);
    expect(payload).toBeTruthy();
    // attacker submits fast and accurate
    const res1 = ms.submitQTE(match.id, a.id, 0.2, payload.word);
    // defender submits slower
    const res2 = ms.submitQTE(match.id, b.id, 0.6, payload.word);

    // After resolution, the match object should have numeric hp values
    const m = ms.getMatch(match.id)!;
    expect(typeof m.a.hp).toBe('number');
    expect(typeof m.b!.hp).toBe('number');
    expect(m.a.hp).toBeGreaterThanOrEqual(0);
    expect(m.b!.hp).toBeGreaterThanOrEqual(0);
  });

  it('both players hp remain numeric when defender parries', () => {
    const { a, b } = makePlayers(ms);
    const match = ms.matchmake(a.id, b.id);
    const payload = ms.startQTE(match.id, a.id);
    // attacker submits slow and inaccurate
    const res1 = ms.submitQTE(match.id, a.id, 0.9, 'wrong');
    // defender submits faster and accurate
    const res2 = ms.submitQTE(match.id, b.id, 0.3, payload.word);

    const m = ms.getMatch(match.id)!;
    expect(typeof m.a.hp).toBe('number');
    expect(typeof m.b!.hp).toBe('number');
  });
});
