import { Injectable } from '@nestjs/common';
import {
  resolveQTE,
  pickRandomWord,
  Player,
  createPlayer,
  applyDamage,
  synchronizePlayers,
} from '@atoyr/shared';
import type { Response } from 'express';

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
  // pending submissions keyed by `${matchId}:${turn}`
  private pendingSubmissions = new Map<
    string,
    {
      turn: number;
      initiatorId?: string;
      times: Map<string, number>;
      texts: Map<string, string | null>;
      word?: string;
    }
  >();
  // Classic SSE: per-player set of open Response objects
  private sseClients = new Map<string, Set<Response>>();
  // pending SSE events for players who are not currently connected
  // store either raw data or a pre-serialized JSON string under `dataStr`
  private pendingEvents = new Map<string, Array<{ event: string; data?: any; dataStr?: string }>>();
  // pending acknowledgements for completed turns: matchId -> { turn, awaiting:Set<playerId>, timer }
  private pendingAcks = new Map<
    string,
    { turn: number; awaiting: Set<string>; timer?: NodeJS.Timeout }
  >();
  // store the last emitted SSE payload per player for debugging
  private lastEmitted = new Map<string, { event: string; json: string }>();

  addSseClient(playerId: string, res: Response) {
    if (!this.sseClients.has(playerId)) this.sseClients.set(playerId, new Set());
    this.sseClients.get(playerId)!.add(res);
    // flush any pending events for this player to the newly connected client
    const pending = this.pendingEvents.get(playerId);
    if (pending && pending.length) {
      const write = (r: Response, ev: { event: string; data?: any; dataStr?: string }) => {
        try {
          const json = typeof ev.dataStr === 'string' ? ev.dataStr : JSON.stringify(ev.data);
          const payload = `event: ${ev.event}\ndata: ${json}\n\n`;
          r.write(payload);
        } catch (e) {}
      };
      for (const ev of pending) {
        try {
          write(res, ev);
        } catch (e) {}
      }
      // once flushed to this connection, drop pending events for the player
      this.pendingEvents.delete(playerId);
    }
  }

  // Ensure player/match objects sent over SSE are plain POJOs with numeric fields
  private sanitizePlayer(p: Player | null | undefined) {
    if (!p) return null;
    return {
      id: p.id,
      name: p.name,
      level: Number(p.level) || 0,
      maxHp: Number(p.maxHp) || 0,
      hp: Number(p.hp) || 0,
      attack: Number(p.attack) || 0,
      defense: Number(p.defense) || 0,
      words: Array.isArray(p.words) ? p.words.slice() : [],
    } as Player;
  }

  private sanitizeMatch(m: Match) {
    return {
      id: m.id,
      turn: m.turn,
      logs: Array.isArray(m.logs) ? m.logs.slice() : [],
      a: this.sanitizePlayer(m.a) as Player,
      b: this.sanitizePlayer(m.b) as Player | null,
    } as Match;
  }

  removeSseClient(playerId: string, res: Response) {
    const set = this.sseClients.get(playerId);
    if (set) {
      set.delete(res);
      if (set.size === 0) this.sseClients.delete(playerId);
    }
  }

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

      // determine weak/strong by level (tie-breaker by id), synchronize strong if needed
      const players = [me, o].sort((p1, p2) => p1.level - p2.level || p1.id.localeCompare(p2.id));
      const weak = players[0];
      const strong = players[1];
      const syncedStrong = synchronizePlayers(weak, strong);

      const a = weak;
      const b = syncedStrong === strong ? strong : syncedStrong;

      const match: Match = { id: `m-${Date.now()}`, a, b, turn: 0, logs: [] };
      // ensure distinct objects for a and b
      if (match.a === match.b) {
        match.b = { ...match.b } as Player;
      }
      this.matches.set(match.id, match);
      // notify both players that match is ready
      this.notifyMatchReady(match);
      try {
        // auto-start a QTE when match is created with two players
        this.startQTE(match.id, me.id);
      } catch (e) {
        // ignore start errors (e.g., no words); match is still valid
      }
      return match;
    }

    // No opponentId: try to join an existing pending match (one where `b` is null)
    const pending = Array.from(this.matches.values()).find((m) => m.b === null && m.a.id !== id);
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
    if (!playerId) return;
    const clients = this.sseClients.get(playerId);
    const json = JSON.stringify(data);
    const payload = `event: ${event}\ndata: ${json}\n\n`;
    // debug: record last emitted payload and log summary
    try {
      this.lastEmitted.set(String(playerId), { event, json });
      // lightweight logging; can be toggled by inspecting server logs
      // eslint-disable-next-line no-console
      console.debug(`SSE emit -> player=${playerId} event=${event} jsonLen=${json.length}`);
    } catch (e) {}
    if (!clients || clients.size === 0) {
      // queue the event for delivery when the player reconnects
      if (!this.pendingEvents.has(playerId)) this.pendingEvents.set(playerId, []);
      // store the pre-serialized JSON so future mutations to `data` don't change what's sent
      this.pendingEvents.get(playerId)!.push({ event, dataStr: json });
      return;
    }

    // If there are queued events for this player, flush them first to preserve order
    const queued = this.pendingEvents.get(playerId);
    if (queued && queued.length) {
      for (const ev of queued) {
        const qjson = typeof ev.dataStr === 'string' ? ev.dataStr : JSON.stringify(ev.data);
        const queuedPayload = `event: ${ev.event}\ndata: ${qjson}\n\n`;
        for (const res of clients) {
          try {
            res.write(queuedPayload);
          } catch (e) {
            // ignore individual write errors
          }
        }
      }
      this.pendingEvents.delete(playerId);
    }

    for (const res of clients) {
      try {
        res.write(payload);
      } catch (e) {
        // ignore
      }
    }
  }

  // debug helper: return last emitted payload for a player if present
  getLastEmitted(playerId: string) {
    return this.lastEmitted.get(playerId) || null;
  }

  notifyMatchReady(match: Match) {
    const a = match.a;
    const b = match.b;
    if (a) {
      this.emitEvent(a.id, 'match_ready', {
        matchId: match.id,
        you: this.sanitizePlayer(a),
        opponent: this.sanitizePlayer(b),
      });
    }
    if (b) {
      this.emitEvent(b.id, 'match_ready', {
        matchId: match.id,
        you: this.sanitizePlayer(b),
        opponent: this.sanitizePlayer(a),
      });
    }
  }

  notifyTurn(
    match: Match,
    q: any,
    applied: number,
    turnNumber?: number,
    messages?: Record<string, string>,
  ) {
    // send turn result to both players (including attacker and defender)
    if (!match) return;
    const a = match.a;
    const b = match.b;
    const t = typeof turnNumber === 'number' ? turnNumber : match.turn;
    // sanitize q to ensure numeric damage is present for non-miss kinds
    const sanitizeQ = (qobj: any) => {
      if (!qobj || typeof qobj !== 'object') return qobj;
      const kind = qobj.kind;
      if (kind === 'hit' || kind === 'critical' || kind === 'parry' || kind === 'block') {
        const raw = Number(qobj.damage);
        const dmg = Number.isFinite(raw) ? Math.round(raw) : 0;
        // ensure at least 1 for non-miss outcomes to avoid zero/NaN damage in logs
        const potential = Math.max(0, dmg);
        const damage = Math.max(1, dmg);
        return { kind, damage, potential };
      }
      return { kind: kind ?? qobj.kind };
    };

    const safeQ = sanitizeQ(q);

    const basePayload = {
      matchId: match.id,
      q: safeQ,
      applied,
      match: this.sanitizeMatch(match),
      turn: t,
    };
    if (a)
      this.emitEvent(a.id, 'turn_result', { ...basePayload, message: messages?.[a.id] ?? null });
    if (b)
      this.emitEvent(b.id, 'turn_result', { ...basePayload, message: messages?.[b.id] ?? null });

    // set up pending ACKs for this turn: wait for both players to ACK before starting next QTE
    if (a && b) {
      const awaiting = new Set<string>([a.id, b.id]);
      // clear existing pendingAcks for this match if any
      const prev = this.pendingAcks.get(match.id);
      if (prev && prev.timer) {
        clearTimeout(prev.timer);
      }
      // set up a fallback timer to auto-start next QTE after 5s if ACKs don't arrive
      const timer = setTimeout(() => {
        try {
          this.pendingAcks.delete(match.id);
          // start next QTE even if not all ACKed
          this.startQTE(match.id);
        } catch (e) {}
      }, 5000);
      this.pendingAcks.set(match.id, { turn: t, awaiting, timer });
    }
  }

  startQTE(matchId: string, initiatorId?: string) {
    const m = this.matches.get(matchId);
    if (!m) throw new Error('match not found');
    if (!m.a || !m.b) throw new Error('match does not have two players');

    // pick a word from the intersection/union of both players' word lists (fall back to either)
    const words = Array.from(new Set([...(m.a.words || []), ...(m.b.words || [])]));
    if (words.length === 0) throw new Error('no words available for QTE');
    const word = pickRandomWord(words);

    // Determine attacker (initiator). If not provided, default to player A.
    const attackerId = initiatorId ?? m.a.id;

    // For each recipient, include their role for the upcoming turn
    const payloadForA = {
      matchId: m.id,
      turn: m.turn,
      word,
      initiatorId: attackerId,
      role: attackerId === m.a.id ? 'attack' : 'defend',
    };
    const payloadForB = {
      matchId: m.id,
      turn: m.turn,
      word,
      initiatorId: attackerId,
      role: attackerId === m.b.id ? 'attack' : 'defend',
    };

    // notify both players to start the QTE with per-player role info
    if (m.a) this.emitEvent(m.a.id, 'qte_start', payloadForA);
    if (m.b) this.emitEvent(m.b.id, 'qte_start', payloadForB);

    // create a pending submissions entry for this turn
    const key = `${m.id}:${m.turn}`;
    this.pendingSubmissions.set(key, {
      turn: m.turn,
      initiatorId,
      times: new Map(),
      texts: new Map(),
      word,
    });

    return { matchId: m.id, turn: m.turn, word, initiatorId: attackerId };
  }

  // Player submits their QTE time for the current turn. If both players have submitted, resolve the turn and notify both.
  submitQTE(matchId: string, playerId: string, time: number | null, text?: string | null) {
    const m = this.matches.get(matchId);
    if (!m) throw new Error('match not found');
    if (!m.a || !m.b) throw new Error('match does not have two players');

    const key = `${m.id}:${m.turn}`;
    let entry = this.pendingSubmissions.get(key);
    if (!entry) {
      // No QTE started for this turn yet
      entry = { turn: m.turn, initiatorId: undefined, times: new Map(), texts: new Map() };
      this.pendingSubmissions.set(key, entry);
    }

    // store the player's submission (time may be null meaning miss)
    if (typeof time === 'number') entry.times.set(playerId, time);
    else entry.times.set(playerId, NaN);
    entry.texts.set(playerId, text ?? null);

    // If both players haven't submitted yet, return pending
    const players = [m.a.id, m.b.id];
    const hasA = entry.times.has(m.a.id);
    const hasB = entry.times.has(m.b.id);
    if (!hasA || !hasB) {
      return { status: 'pending', turn: entry.turn };
    }

    // Both submitted: determine attacker/defender
    const initiatorId = entry.initiatorId || undefined;
    let attackerId = initiatorId;
    if (!attackerId) {
      // fallback: first submitter in the Map
      const first = entry.times.keys().next();
      attackerId = first.done ? m.a.id : first.value;
    }

    const defenderId = attackerId === m.a.id ? m.b.id : m.a.id;
    const attacker = attackerId === m.a.id ? m.a : m.b;
    const defender = defenderId === m.a.id ? m.a : m.b;

    const attackerTime = Number(entry.times.get(attackerId));
    const defenderTime = Number(entry.times.get(defenderId));
    const attackerText = entry.texts.get(attackerId) ?? null;
    const defenderText = entry.texts.get(defenderId) ?? null;

    const q = resolveQTE(
      { time: Number.isFinite(attackerTime) ? attackerTime : null, text: attackerText },
      { time: Number.isFinite(defenderTime) ? defenderTime : null, text: defenderText },
      entry.word ?? null,
      attacker.attack,
    );

    let applied = 0;
    // craft personalized messages for both players and server logs
    let msgA = '';
    let msgB = '';
    if (q.kind === 'miss') {
      msgA = attacker.id === m.a.id ? 'You missed' : 'Your attack missed';
      msgB = attacker.id === m.a.id ? 'Your attack missed' : 'You missed';
      m.logs.unshift(`${attacker.id} missed`);
    } else if (q.kind === 'critical' || q.kind === 'hit') {
      const dmg = Number((q as any).damage) || 0;
      applied = applyDamage(defender, dmg);
      msgA = attacker.id === m.a.id ? `You deal damage: ${applied}` : `You take damage: ${applied}`;
      msgB = attacker.id === m.a.id ? `You take damage: ${applied}` : `You deal damage: ${applied}`;
      m.logs.unshift(`${attacker.id} hit ${defender.id} for ${applied} (${q.kind})`);
    } else if (q.kind === 'block') {
      const partial = Math.round((Number((q as any).damage) || 0) * 0.25);
      applied = applyDamage(defender, partial);
      msgA =
        attacker.id === m.a.id
          ? `You deal partial damage: ${applied}`
          : `You take partial damage: ${applied}`;
      msgB =
        attacker.id === m.a.id
          ? `You take partial damage: ${applied}`
          : `You deal partial damage: ${applied}`;
      m.logs.unshift(`${attacker.id} partially hit ${defender.id} for ${applied}`);
    } else if (q.kind === 'parry') {
      const dmgParry = Number((q as any).damage) || 0;
      applied = applyDamage(attacker, dmgParry);
      msgA =
        attacker.id === m.a.id
          ? `Your attack was parried. You take ${applied}`
          : `You parried and countered for ${applied}`;
      msgB =
        attacker.id === m.a.id
          ? `You parried and countered for ${applied}`
          : `Your attack was parried. You take ${applied}`;
      m.logs.unshift(`${defender.id} parried and countered ${attacker.id} for ${applied}`);
    }

    // notify subscribers about the completed turn first (do not increment m.turn yet)
    try {
      // notifyTurn will send the q and applied; include per-player messages as part of the payload
      this.notifyTurn(m, q, applied, undefined, { [m.a.id]: msgA, [m.b.id!]: msgB });
    } catch (e) {
      // ignore notify errors
    }

    // increment turn after notifying about the completed turn
    const completedTurn = m.turn;
    m.turn += 1;

    // cleanup pending entry
    this.pendingSubmissions.delete(key);

    // Next QTE start is controlled by ACKs (notifyTurn sets up pendingAcks)
    // or its fallback timer. Do NOT auto-start another QTE here to avoid
    // emitting duplicate `qte_start` events (one from this code path and
    // another from the ACK handler or fallback).

    // Important: do NOT return the resolved payload here. Clients must rely on SSE
    // for the authoritative turn_result and qte_start ordering. Returning the
    // result in the HTTP response risks clients observing the result before the
    // server has emitted turn_result (race with SSE).
    return { status: 'resolved', turn: completedTurn };
  }

  getMatch(id: string) {
    return this.matches.get(id) || null;
  }

  // Called when a client ACKs a turn_result for a given match/turn
  ackTurn(matchId: string, playerId: string, turn: number) {
    const entry = this.pendingAcks.get(matchId);
    if (!entry) return { status: 'no-pending' };
    if (entry.turn !== turn) return { status: 'mismatch', expected: entry.turn };
    entry.awaiting.delete(playerId);
    if (entry.awaiting.size === 0) {
      // all ACKs received; clear timer and start next QTE
      if (entry.timer) clearTimeout(entry.timer);
      this.pendingAcks.delete(matchId);
      try {
        this.startQTE(matchId);
      } catch (e) {}
      return { status: 'started' };
    }
    return { status: 'waiting', remaining: entry.awaiting.size };
  }

  joinMatch(matchId: string, playerId: string) {
    const m = this.matches.get(matchId);
    if (!m) throw new Error('match not found');
    if (m.b !== null) throw new Error('match already has two players');

    const me = this.players.get(playerId);
    if (!me) throw new Error('player not found');

    const other = m.a;

    // determine weak/strong by level (tie-breaker by id), synchronize strong if needed
    const players = [other, me].sort((p1, p2) => p1.level - p2.level || p1.id.localeCompare(p2.id));
    const weak = players[0];
    const strong = players[1];
    const syncedStrong = synchronizePlayers(weak, strong);

    m.a = weak;
    m.b = syncedStrong === strong ? strong : syncedStrong;
    // ensure distinct objects
    if (m.a === m.b) {
      m.b = { ...m.b } as Player;
    }
    this.matches.set(m.id, m);
    this.notifyMatchReady(m);
    try {
      // the player who joined (playerId) triggered the match to start
      this.startQTE(m.id, playerId);
    } catch (e) {
      // ignore
    }
    return m;
  }

  resolveTurn(matchId: string, attackerId: string, attackerTime?: number, defenderTime?: number) {
    const m = this.matches.get(matchId);
    if (!m) throw new Error('match not found');
    // Find attacker by id first, then try by name as a fallback (clients may send name)
    let attacker: Player | null = null;
    if (attackerId && m.a && m.a.id === attackerId) attacker = m.a;
    else if (attackerId && m.b && m.b.id === attackerId) attacker = m.b;
    // fallback: maybe client sent the player's name instead of id
    else if (attackerId && m.a && m.a.name === attackerId) attacker = m.a;
    else if (attackerId && m.b && m.b.name === attackerId) attacker = m.b;

    const defender = attacker === m.a ? m.b : m.a;
    if (!attacker) {
      const aId = m.a ? `${m.a.id}${m.a.name ? ` (${m.a.name})` : ''}` : 'none';
      const bId = m.b ? `${m.b.id}${m.b.name ? ` (${m.b.name})` : ''}` : 'none';
      throw new Error(
        `attacker not in match (attackerId=${String(attackerId)}, matchPlayers=${aId},${bId})`,
      );
    }
    if (!defender) throw new Error('defender not in match or match incomplete');

    const q = resolveQTE(
      { time: typeof attackerTime === 'number' ? attackerTime : null },
      { time: typeof defenderTime === 'number' ? defenderTime : null },
      undefined,
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
