import React, { useEffect, useRef, useState } from 'react';

type ResolveResponse = { result?: { kind: string; damage?: number }; q?: { kind: string; damage?: number }; match?: any; applied?: number };

export default function App() {
  const [name, setName] = useState('player1');
  const [playerId, setPlayerId] = useState<string | null>(null);
  const [words, setWords] = useState<string[]>(['apple', 'banana', 'cherry', 'date', 'elder']);
  const [opponentWords, setOpponentWords] = useState<string[]>([]);
  const [matchId, setMatchId] = useState<string | null>(null);
  const [log, setLog] = useState<string[]>([]);
  const [roundWord, setRoundWord] = useState<string | null>(null);
  const [roundIndex, setRoundIndex] = useState(0);
  const startRef = useRef<number | null>(null);

  useEffect(() => {
    // cleanup on unmount
    return () => {
      // nothing to cleanup here; EventSource handled via ref when created
    };
  }, []);

  // keep a ref to EventSource so we can close it on unmount or player change
  const esRef = useRef<EventSource | null>(null);
  function initSSE(forPlayerId: string) {
    if (!forPlayerId) return;
    // close previous
    if (esRef.current) {
      try {
        esRef.current.close();
      } catch (e) { }
      esRef.current = null;
    }

    const url = `/api/events/${encodeURIComponent(forPlayerId)}`;
    const es = new EventSource(url);
    esRef.current = es;

    es.onopen = () => {
      setLog((l) => [`SSE connected for ${forPlayerId}`, ...l]);
    };

    es.addEventListener('match_ready', (ev: any) => {
      try {
        const payload = JSON.parse((ev as MessageEvent).data);
        setMatchId(payload.matchId);
        const opponent = payload.opponent;
        const oppWords: string[] = (opponent && Array.isArray(opponent.words) && opponent.words) || [];
        setOpponentWords(oppWords.slice(0, 5));
        setLog((l) => [`Match ready: opponent=${opponent?.id ?? 'unknown'}`, ...l]);
      } catch (e) {
        setLog((l) => [`match_ready parse error: ${(e as Error).message}`, ...l]);
      }
    });

    es.addEventListener('turn_result', (ev: any) => {
      try {
        const payload = JSON.parse((ev as MessageEvent).data);
        const q = payload.q;
        const applied = payload.applied;
        setLog((l) => [`Turn result: ${q?.kind ?? 'unknown'} dmg=${q?.damage ?? applied ?? 0}`, ...l]);
        // update local match/opponent state if present
        if (payload.match) {
          const opponent = payload.match.a?.id === forPlayerId ? payload.match.b : payload.match.a;
          const oppWords: string[] = (opponent && Array.isArray(opponent.words) && opponent.words) || [];
          setOpponentWords(oppWords.slice(0, 5));
        }
      } catch (e) {
        setLog((l) => [`turn_result parse error: ${(e as Error).message}`, ...l]);
      }
    });

    es.onerror = (err) => {
      setLog((l) => [`SSE error: ${String(err)}`, ...l]);
    };

    return es;
  }

  useEffect(() => {
    if (!playerId) return;
    const es = initSSE(playerId);
    return () => {
      try {
        es && es.close();
      } catch (e) { }
      esRef.current = null;
    };
  }, [playerId]);

  async function createPlayer() {
    try {
      const id = name;
      const res = await fetch('/api/player', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id, name }),
      });
      const data = await res.json();
      setPlayerId(id);
      setLog((l) => [`Created player ${name}`, ...l]);
      // if server returned player object with words, update local words
      if (data && Array.isArray(data.words) && data.words.length) setWords(data.words.slice(0, 5));
    } catch (err) {
      setLog((l) => [`createPlayer error: ${(err as Error).message}`, ...l]);
    }
  }

  async function matchmake() {
    try {
      const id = playerId || name;
      // ensure SSE is open for this player so we receive match_ready notifications
      if (!playerId) {
        setPlayerId(id);
        try {
          initSSE(id);
          setLog((l) => [`SSE initiated for ${id}`, ...l]);
        } catch (e) {
          setLog((l) => [`SSE init error: ${(e as Error).message}`, ...l]);
        }
      }
      // ensure the player is registered on the server so matchmake can find them
      try {
        await fetch('/api/player', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ id, name }),
        });
      } catch (e) {
        // ignore - server may already have the player or will error; we'll handle below
      }
      const res = await fetch('/api/matchmake', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id }),
      });
      const match = await res.json();
      if (!match || !match.id) {
        setLog((l) => [`matchmake: invalid response`, ...l]);
        return;
      }
      setMatchId(match.id);
      // find opponent (the other player)
      const opponent = match.a?.id === id ? match.b : match.a;
      const oppWords: string[] = (opponent && Array.isArray(opponent.words) && opponent.words) || [];
      setOpponentWords(oppWords.slice(0, 5));
      setLog((l) => [`Matched vs opponent with words: ${oppWords.join(', ')}`, ...l]);
    } catch (err) {
      setLog((l) => [`matchmake error: ${(err as Error).message}`, ...l]);
    }
  }

  function startRound() {
    if (opponentWords.length === 0) return;
    const w = opponentWords[Math.floor(Math.random() * opponentWords.length)];
    setRoundWord(w);
    setRoundIndex((i) => i + 1);
    startRef.current = performance.now();
  }

  async function submitAnswer(answer: string) {
    const end = performance.now();
    const elapsed = startRef.current ? (end - startRef.current) / 1000 : null;
    startRef.current = null;

    // simulate opponent time (random)
    const opponentTime = Math.random() * 1.2; // seconds

    if (!matchId) {
      setLog((l) => [`No active match; cannot submit answer`, ...l]);
      return;
    }

    const payload = {
      matchId,
      attackerId: playerId || name,
      attackerTime: elapsed,
      defenderTime: opponentTime,
    };

    try {
      const res = await fetch('/api/resolve', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      const data: ResolveResponse = await res.json();
      // server previously returns { q, match, applied }
      const q = data.q || data.result;
      if (!q) {
        setLog((l) => [`Round ${roundIndex}: resolve returned unexpected response: ${JSON.stringify(data)}`, ...l]);
      } else {
        setLog((l) => [`Round ${roundIndex}: answer='${answer}' elapsed=${elapsed?.toFixed(3)}s vs opp=${opponentTime.toFixed(3)}s -> ${q.kind} dmg=${q.damage ?? 0}`, ...l]);
      }
    } catch (err) {
      setLog((l) => [`resolve error: ${(err as Error).message}`, ...l]);
    }
    setRoundWord(null);
  }

  return (
    <div className="app">
      <h1>A Toy R — Integrated Prototype</h1>

      <div>
        <label>Player name:</label>
        <input value={name} onChange={(e) => setName(e.target.value)} />
        <button onClick={createPlayer}>Create Player (local)</button>
      </div>

      <div>
        <label>My words (5):</label>
        <ul>
          {words.map((w) => (
            <li key={w}>{w}</li>
          ))}
        </ul>
      </div>

      <div>
        <button onClick={matchmake}>Matchmake (call server)</button>
        <div>Opponent words: {opponentWords.join(', ')}</div>
      </div>

      <div>
        <button onClick={startRound} disabled={!opponentWords.length || !!roundWord}>
          Start QTE Round
        </button>
        {roundWord && (
          <div>
            <div>Type this word as fast as you can:</div>
            <h2>{roundWord}</h2>
            <input
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  const val = (e.target as HTMLInputElement).value;
                  (e.target as HTMLInputElement).value = '';
                  submitAnswer(val);
                }
              }}
            />
          </div>
        )}
      </div>

      <div className="log">
        {log.map((l, i) => (
          <div key={i}>{l}</div>
        ))}
      </div>
    </div>
  );
}
