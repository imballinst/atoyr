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
import { Subject, Observable } from 'rxjs';
import { filter as rxFilter, map as rxMap } from 'rxjs/operators';

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
  // event stream for SSE (emits raw events { playerId?, event, data })
  private events$ = new Subject<{ playerId?: string; event: string; data: any }>();

  // public observable for controllers to subscribe and filter
  public eventStream: Observable<{ playerId?: string; event: string; data: any }> =
    this.events$.asObservable();

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
      // ensure distinct objects for a and b
      if (match.a === match.b) {
        match.b = { ...match.b } as Player;
      }
      this.matches.set(match.id, match);
      // notify both players that match is ready
      this.notifyMatchReady(match);
      return match;
    }

    // No opponentId: try to join an existing pending match (one where `b` is null)
    const pending = Array.from(this.matches.values()).find((m) => m.b === null && m.a.id !== id);
    console.info(pending);
    if (pending) {
      // delegate to joinMatch for atomic/safe join logic
      return this.joinMatch(pending.id, id);
    }

    // Otherwise create a pending match (waiting for a second player)
    const match: Match = { id: `m-${Date.now()}`, a: me, b: null, turn: 0, logs: [] };
    this.matches.set(match.id, match);
    return match;
  }

  private emitEvent(playerId: string | undefined, event: string, data: any) {
    console.info(playerId, event, data);
    this.events$.next({ playerId, event, data });
  }

  notifyMatchReady(match: Match) {
    const a = match.a;
    const b = match.b;
    if (a) {
      this.emitEvent(a.id, 'match_ready', { matchId: match.id, you: a, opponent: b });
    }
    if (b) {
      this.emitEvent(b.id, 'match_ready', { matchId: match.id, you: b, opponent: a });
    }
  }

  notifyTurn(match: Match, q: any, applied: number) {
    // send turn result to both players (including attacker and defender)
    if (!match) return;
    const a = match.a;
    const b = match.b;
    if (a) this.emitEvent(a.id, 'turn_result', { matchId: match.id, q, applied, match });
    if (b) this.emitEvent(b.id, 'turn_result', { matchId: match.id, q, applied, match });
  }

  getMatch(id: string) {
    return this.matches.get(id) || null;
  }

  joinMatch(matchId: string, playerId: string) {
    const m = this.matches.get(matchId);
    if (!m) throw new Error('match not found');
    if (m.b !== null) throw new Error('match already has two players');

    const me = this.players.get(playerId);
    if (!me) throw new Error('player not found');

    const other = m.a;

    // synchronize levels if needed
    const weak = other.level < me.level ? other : me;
    const strong = other.level < me.level ? me : other;
    const syncedStrong = synchronizePlayers(weak, strong);

    const a = other.level <= me.level ? other : syncedStrong;
    const b = other.level <= me.level ? syncedStrong : other;

    m.a = a;
    m.b = b;
    // ensure distinct objects
    if (m.a === m.b) {
      m.b = { ...m.b } as Player;
    }
    this.matches.set(m.id, m);
    this.notifyMatchReady(m);
    return m;
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
    // notify subscribers about the turn result
    try {
      this.notifyTurn(m, q, applied);
    } catch (e) {
      // ignore notify errors
    }
    return { match: m, q, applied };
  }
}
