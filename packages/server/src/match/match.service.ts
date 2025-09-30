import { Injectable } from '@nestjs/common';
import {
  resolveQTE,
  pickRandomWord,
  Player,
  createPlayer,
  applyDamage,
  synchronizePlayers,
  statsForLevel,
} from '@atoyr/shared';

type Match = {
  id: string;
  a: Player;
  b: Player | null;
  turn: number;
  logs: string[];
};

@Injectable()
export class MatchService {
  private players = new Map<string, Player>();
  private matches = new Map<string, Match>();

  root() {
    return 'A Toy R server (NestJS prototype)';
  }

  createPlayer(id: string, name?: string, level = 1) {
    const p = createPlayer(id, name, level, ['quick', 'apple', 'banana', 'cherry', 'delta']);
    this.players.set(id, p);
    return p;
  }

  getPlayer(id: string) {
    return this.players.get(id) || null;
  }

  matchmake(id: string, opponentId?: string) {
    const me = this.players.get(id);
    if (!me) throw new Error('player not found');
    // If an opponentId was provided, create a match directly between the two players
    if (opponentId) {
      const o = this.players.get(opponentId);
      if (!o) throw new Error('opponent not found');

      // synchronize if huge level difference
      const weak = me.level < o.level ? me : o;
      const strong = me.level < o.level ? o : me;
      const syncedStrong = synchronizePlayers(weak, strong);

      const a = me.level <= o.level ? me : syncedStrong;
      const b = me.level <= o.level ? syncedStrong : me;

      const match: Match = { id: `m-${Date.now()}`, a, b, turn: 0, logs: [] };
      this.matches.set(match.id, match);
      return match;
    }

    // No opponentId: try to join an existing pending match (one where `b` is null)
    const pending = Array.from(this.matches.values()).find((m) => m.b === null && m.a.id !== id);
    if (pending) {
      // join the pending match as player b
      const other = pending.a;

      // synchronize levels if needed
      const weak = other.level < me.level ? other : me;
      const strong = other.level < me.level ? me : other;
      const syncedStrong = synchronizePlayers(weak, strong);

      const a = other.level <= me.level ? other : syncedStrong;
      const b = other.level <= me.level ? syncedStrong : other;

      pending.a = a;
      pending.b = b;
      this.matches.set(pending.id, pending);
      return pending;
    }

    // Otherwise create a pending match (waiting for a second player)
    const match: Match = { id: `m-${Date.now()}`, a: me, b: null, turn: 0, logs: [] };
    this.matches.set(match.id, match);
    return match;
  }

  getMatch(id: string) {
    return this.matches.get(id) || null;
  }

  resolveTurn(matchId: string, attackerId: string, attackerTime?: number, defenderTime?: number) {
    const m = this.matches.get(matchId);
    if (!m) throw new Error('match not found');
    const attacker = m.a.id === attackerId ? m.a : m.b && m.b.id === attackerId ? m.b : null;
    const defender = attacker === m.a ? m.b : m.a;
    if (!attacker) throw new Error('attacker not in match');
    if (!defender) throw new Error('defender not in match or match incomplete');

    const q = resolveQTE(
      typeof attackerTime === 'number' ? attackerTime : null,
      typeof defenderTime === 'number' ? defenderTime : null,
      attacker.attack,
    );

    let applied = 0;
    if (q.kind === 'miss') {
      m.logs.unshift(`${attacker.id} missed`);
    } else if (q.kind === 'critical' || q.kind === 'hit') {
      applied = applyDamage(defender, q.damage);
      m.logs.unshift(`${attacker.id} hit ${defender.id} for ${applied} (${q.kind})`);
    } else if (q.kind === 'block') {
      const partial = Math.round(q.damage * 0.25);
      applied = applyDamage(defender, partial);
      m.logs.unshift(`${attacker.id} partially hit ${defender.id} for ${applied}`);
    } else if (q.kind === 'parry') {
      // parry counter
      applied = applyDamage(attacker, q.damage);
      m.logs.unshift(`${defender.id} parried and countered ${attacker.id} for ${applied}`);
    }

    m.turn += 1;
    return { match: m, q, applied };
  }
}
