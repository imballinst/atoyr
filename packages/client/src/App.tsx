import React, { useEffect, useRef, useState } from 'react';

type ResolveResponse = {
  result?: { kind: string; damage?: number };
  q?: { kind: string; damage?: number };
  match?: any;
  applied?: number;
};

type Player = {
  id: string;
  name?: string;
  level: number;
  maxHp: number;
  hp: number;
  attack?: number;
  defense?: number;
  words?: string[];
};

export default function App() {
  const [name, setName] = useState('player1');
  const [playerId, setPlayerId] = useState<string | null>(null);
  const [player, setPlayer] = useState<Player | null>(null);
  const [words, setWords] = useState<string[]>(['apple', 'banana', 'cherry', 'date', 'elder']);
  const [opponentWords, setOpponentWords] = useState<string[]>([]);
  const [opponent, setOpponent] = useState<Player | null>(null);
  const [matchId, setMatchId] = useState<string | null>(null);
  const [log, setLog] = useState<string[]>([]);
  const [roundWord, setRoundWord] = useState<string | null>(null);
  const [roundIndex, setRoundIndex] = useState(0);
  const [waitingForResult, setWaitingForResult] = useState(false);
  const startRef = useRef<number | null>(null);

  useEffect(() => {
    const name = sessionStorage.getItem('name');
    if (name) setName(name);
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
        const you: Player | null = payload.you || null;
        const opp: Player | null = payload.opponent || null;
        if (you) setPlayer(you);
        if (opp) setOpponent(opp);
        const oppWords: string[] = (opp && Array.isArray(opp.words) && opp.words) || [];
        setOpponentWords(oppWords.slice(0, 5));
        setLog((l) => [`Match ready: opponent=${opp?.id ?? 'unknown'}`, ...l]);
      } catch (e) {
        setLog((l) => [`match_ready parse error: ${(e as Error).message}`, ...l]);
      }
    });

    es.addEventListener('turn_result', (ev: any) => {
      try {
        const payload = JSON.parse((ev as MessageEvent).data);
        const q = payload.q;
        const applied = payload.applied;
        setLog((l) => [
          `Turn result: ${q?.kind ?? 'unknown'} dmg=${q?.damage ?? applied ?? 0}`,
          ...l,
        ]);
        // update local match/opponent state if present
        if (payload.match) {
          const me = payload.match.a?.id === forPlayerId ? payload.match.a : payload.match.b;
          const opponentObj = payload.match.a?.id === forPlayerId ? payload.match.b : payload.match.a;
          if (me) setPlayer(me);
          if (opponentObj) setOpponent(opponentObj);
          const oppWords: string[] = (opponentObj && Array.isArray(opponentObj.words) && opponentObj.words) || [];
          setOpponentWords(oppWords.slice(0, 5));
        }
        // clear waiting state and current round when server sends the final result
        setWaitingForResult(false);
        setRoundWord(null);
      } catch (e) {
        setLog((l) => [`turn_result parse error: ${(e as Error).message}`, ...l]);
      }
    });

    es.addEventListener('qte_start', (ev: any) => {
      try {
        const payload = JSON.parse((ev as MessageEvent).data);
        // payload: { matchId, turn, word, initiatorId }
        setMatchId(payload.matchId);
        setRoundWord(payload.word);
        setRoundIndex((i) => i + 1);
        startRef.current = performance.now();
        setLog((l) => [`QTE started: word='${payload.word}' turn=${payload.turn} initiator=${payload.initiatorId}`, ...l]);
      } catch (e) {
        setLog((l) => [`qte_start parse error: ${(e as Error).message}`, ...l]);
      }
    });

    es.onerror = (err) => {
      console.error(err);
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
      if (data && data.id) setPlayer(data as Player);
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
      const oppWords: string[] =
        (opponent && Array.isArray(opponent.words) && opponent.words) || [];
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

  async function initiateQteServer() {
    if (!matchId) {
      setLog((l) => [`No match to start QTE on`, ...l]);
      return;
    }
    try {
      const res = await fetch(`/api/match/${encodeURIComponent(matchId)}/start`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ initiatorId: playerId || name }),
      });
      const data = await res.json();
      setLog((l) => [`Requested server QTE start: ${JSON.stringify(data)}`, ...l]);
    } catch (e) {
      setLog((l) => [`initiateQteServer error: ${(e as Error).message}`, ...l]);
    }
  }

  async function submitAnswer(answer: string) {
    const end = performance.now();
    const elapsed = startRef.current ? (end - startRef.current) / 1000 : null;
    startRef.current = null;

    if (!matchId) {
      setLog((l) => [`No active match; cannot submit answer`, ...l]);
      return;
    }

    try {
      const res = await fetch(`/api/match/${encodeURIComponent(matchId)}/submit`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ playerId: playerId || name, time: elapsed }),
      });
      const data = await res.json();
      if (data && data.status === 'pending') {
        setWaitingForResult(true);
        setLog((l) => [`Round ${roundIndex}: submitted, waiting for opponent`, ...l]);
      } else if (data && (data.q || data.match)) {
        // server resolved immediately (both submissions present)
        const q = data.q || data.result;
        setLog((l) => [`Round ${roundIndex}: result immediate: ${q?.kind ?? 'unknown'}`, ...l]);
        setRoundWord(null);
        setWaitingForResult(false);
      } else {
        setLog((l) => [`Round ${roundIndex}: unexpected submit response: ${JSON.stringify(data)}`, ...l]);
      }
    } catch (err) {
      setLog((l) => [`submit error: ${(err as Error).message}`, ...l]);
    }
  }

  return (
    <div className="app">
      <h1>A Toy R — Integrated Prototype</h1>

      {/* HP display */}
      <div style={{ display: 'flex', gap: 24, alignItems: 'center', marginBottom: 16 }}>
        <div style={{ flex: 1 }}>
          <div style={{ fontWeight: 'bold', marginBottom: 6 }}>You</div>
          <div style={{ background: '#333', borderRadius: 8, padding: 4 }}>
            <div
              style={{
                height: 24,
                borderRadius: 6,
                background: 'linear-gradient(90deg,#4caf50,#8bc34a)',
                width: player ? `${Math.max(0, (player.hp / player.maxHp) * 100)}%` : '0%',
                transition: 'width 300ms ease',
              }}
            />
          </div>
          <div style={{ marginTop: 6, fontSize: 14 }}>
            {player ? `${player.hp} / ${player.maxHp} HP` : 'No player'}
          </div>
        </div>

        <div style={{ flex: 1 }}>
          <div style={{ fontWeight: 'bold', marginBottom: 6 }}>Opponent</div>
          <div style={{ background: '#333', borderRadius: 8, padding: 4 }}>
            <div
              style={{
                height: 24,
                borderRadius: 6,
                background: 'linear-gradient(90deg,#f44336,#ff7961)',
                width: opponent ? `${Math.max(0, (opponent.hp / opponent.maxHp) * 100)}%` : '0%',
                transition: 'width 300ms ease',
              }}
            />
          </div>
          <div style={{ marginTop: 6, fontSize: 14 }}>
            {opponent ? `${opponent.hp} / ${opponent.maxHp} HP` : 'No opponent'}
          </div>
        </div>
      </div>

      <div>
        <label>Player name:</label>
        <input value={name} onChange={(e) => {
          setName(e.target.value)
          sessionStorage.setItem('name', e.target.value);
        }} />
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
