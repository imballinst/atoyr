import { useEffect, useRef, useState } from 'react';

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
  const waitingTimerRef = useRef<number | null>(null);
  const [currentTurn, setCurrentTurn] = useState<number | null>(null);
  const [roundCompleted, setRoundCompleted] = useState(false);
  const [lastResult, setLastResult] = useState<string | null>(null);
  const startRef = useRef<number | null>(null);
  const [answerInput, setAnswerInput] = useState('');
  const inputRef = useRef<HTMLInputElement | null>(null);

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
      } catch (e) {}
      esRef.current = null;
    }

    const url = `/api/events/${encodeURIComponent(forPlayerId)}`;
    const es = new EventSource(url);
    esRef.current = es;

    es.onopen = () => {
      setLog((l) => [`SSE connected for ${forPlayerId}`, ...l]);
    };

    es.addEventListener('message', (ev: any) => {
      // generic message event - log it
      setLog((l) => [`SSE message: ${(ev as MessageEvent).data}`, ...l]);
    });

    es.addEventListener('match_ready', (ev: any) => {
      try {
        const payload = JSON.parse((ev as MessageEvent).data);
        setMatchId(payload.matchId);
        const you: Player | null = payload.you || null;
        const opp: Player | null = payload.opponent || null;
        if (you) setPlayer(you);
        if (opp) setOpponent(opp);
        // set current turn if present
        if (payload.turn !== undefined && payload.turn !== null) setCurrentTurn(payload.turn);
        const oppWords: string[] = (opp && Array.isArray(opp.words) && opp.words) || [];
        setOpponentWords(oppWords.slice(0, 5));
        // ensure input is enabled when the match becomes ready
        setWaitingForResult(false);
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
        // log both the generic result and any personalized message
        setLog((l) => [
          `Turn result: ${q?.kind ?? 'unknown'} dmg=${q?.damage ?? applied ?? 0}`,
          ...(payload.message ? [`Message: ${payload.message}`] : []),
          ...l,
        ]);
        // update local match/opponent state if present
        if (payload.match) {
          const me = payload.match.a?.id === forPlayerId ? payload.match.a : payload.match.b;
          const opponentObj =
            payload.match.a?.id === forPlayerId ? payload.match.b : payload.match.a;
          if (me) setPlayer(me);
          if (opponentObj) setOpponent(opponentObj);
          const oppWords: string[] =
            (opponentObj && Array.isArray(opponentObj.words) && opponentObj.words) || [];
          setOpponentWords(oppWords.slice(0, 5));
        }
        // clear waiting state and mark round completed; keep the target visible until next qte_start
        setWaitingForResult(false);
        // clear any fallback waiting timer
        try {
          if (waitingTimerRef.current) {
            window.clearTimeout(waitingTimerRef.current);
            waitingTimerRef.current = null;
          }
        } catch (e) {}
        setRoundCompleted(true);
        // if server sent a personalized message, prefer it; otherwise show generic
        setLastResult(
          payload.message ?? `${q?.kind ?? 'unknown'} dmg=${q?.damage ?? applied ?? 0}`,
        );

        // send ACK back to server indicating we've processed this turn_result
        (async () => {
          try {
            await fetch(`/api/events/${encodeURIComponent(forPlayerId)}/ack`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ matchId: payload.matchId, turn: payload.turn }),
            });
          } catch (e) {
            // ignore ack errors; server has a fallback timer
          }
        })();
      } catch (e) {
        setLog((l) => [`turn_result parse error: ${(e as Error).message}`, ...l]);
      }
    });

    es.addEventListener('qte_start', (ev: any) => {
      try {
        const payload = JSON.parse((ev as MessageEvent).data);
        // payload: { matchId, turn, word, initiatorId }
        setMatchId(payload.matchId);
        // new round starting: clear completed flag and set new target
        setRoundCompleted(false);
        // ensure any previous waiting state is cleared so input is enabled for the new round
        setWaitingForResult(false);
        try {
          if (waitingTimerRef.current) {
            window.clearTimeout(waitingTimerRef.current);
            waitingTimerRef.current = null;
          }
        } catch (e) {}
        setLastResult(null);
        setRoundWord(payload.word);
        setRoundIndex((i) => i + 1);
        setCurrentTurn(payload.turn);
        startRef.current = performance.now();
        // focus the input so the player can continue typing immediately
        setTimeout(() => {
          try {
            inputRef.current && inputRef.current.focus();
          } catch (e) {}
        }, 50);
        setLog((l) => [
          `QTE started: word='${payload.word}' turn=${payload.turn} initiator=${payload.initiatorId}`,
          ...l,
        ]);
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
      } catch (e) {}
      esRef.current = null;
      // clear any waiting timer when component unmounts or player changes
      try {
        if (waitingTimerRef.current) {
          window.clearTimeout(waitingTimerRef.current);
          waitingTimerRef.current = null;
        }
      } catch (e) {}
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

  // local manual round start removed: QTEs are server-driven via SSE

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
        body: JSON.stringify({ playerId: playerId || name, time: elapsed, text: answer }),
      });
      const data = await res.json();
      if (data && data.status === 'pending') {
        setWaitingForResult(true);
        try {
          if (waitingTimerRef.current) window.clearTimeout(waitingTimerRef.current);
        } catch (e) {}
        // fallback: clear waiting state after 8s if no turn_result arrives
        waitingTimerRef.current = window.setTimeout(() => {
          setWaitingForResult(false);
          waitingTimerRef.current = null;
          setLog((l) => [
            `Cleared waiting state (timeout), if you still see issues reconnect or check server logs`,
            ...l,
          ]);
        }, 8000) as unknown as number;
        setLog((l) => [`Turn ${currentTurn ?? roundIndex}: submitted, waiting for opponent`, ...l]);
      } else if (data && data.status === 'resolved') {
        // server resolved the turn synchronously; still wait for authoritative SSE
        setWaitingForResult(true);
        try {
          if (waitingTimerRef.current) window.clearTimeout(waitingTimerRef.current);
        } catch (e) {}
        waitingTimerRef.current = window.setTimeout(() => {
          setWaitingForResult(false);
          waitingTimerRef.current = null;
          setLog((l) => [`Cleared waiting state after resolved response (timeout)`, ...l]);
        }, 8000) as unknown as number;
        setLog((l) => [
          `Turn ${data.turn ?? currentTurn ?? roundIndex}: server resolved, waiting for authoritative result via SSE`,
          ...l,
        ]);
      } else if (data && (data.q || data.match)) {
        // server resolved immediately (both submissions present).
        // Still wait for the authoritative SSE 'turn_result' so client-side state
        // (roundWord/next qte_start) is driven by server events and ordering stays correct.
        const q = data.q || data.result;
        setLog((l) => [
          `Turn ${data.turn ?? currentTurn ?? roundIndex}: result immediate (server resolved): ${q?.kind ?? 'unknown'}`,
          ...l,
        ]);
        // do NOT clear roundWord here; wait for SSE to deliver turn_result and qte_start
        setWaitingForResult(true);
        try {
          if (waitingTimerRef.current) window.clearTimeout(waitingTimerRef.current);
        } catch (e) {}
        waitingTimerRef.current = window.setTimeout(() => {
          setWaitingForResult(false);
          waitingTimerRef.current = null;
          setLog((l) => [`Cleared waiting state after immediate result (timeout)`, ...l]);
        }, 8000) as unknown as number;
      } else {
        setLog((l) => [
          `Round ${roundIndex}: unexpected submit response: ${JSON.stringify(data)}`,
          ...l,
        ]);
      }
    } catch (err) {
      setLog((l) => [`submit error: ${(err as Error).message}`, ...l]);
    }
  }

  return (
    <div className="app">
      <h1>A Toy R — Integrated Prototype</h1>

      {/* Turn indicator */}
      <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 8 }}>
        <div
          style={{
            fontWeight: '700',
            fontSize: 20,
            padding: '6px 12px',
            background: '#222',
            color: '#fff',
            borderRadius: 8,
          }}
        >
          Turn: {currentTurn !== null ? currentTurn : '-'}
        </div>
      </div>

      {/* HP display */}
      <div style={{ display: 'flex', gap: 24, alignItems: 'center', marginBottom: 16 }}>
        <div style={{ flex: 1 }}>
          <div style={{ fontWeight: 'bold', marginBottom: 6 }}>
            {player?.name ?? player?.id ?? 'You'}
          </div>
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
          <div style={{ fontWeight: 'bold', marginBottom: 6 }}>
            {opponent?.name ?? opponent?.id ?? 'Opponent'}
          </div>
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
        <input
          value={name}
          onChange={(e) => {
            setName(e.target.value);
            sessionStorage.setItem('name', e.target.value);
          }}
        />
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
        <div style={{ marginTop: 12 }}>
          <div style={{ marginBottom: 6 }}>
            <div style={{ fontSize: 14, color: '#666' }}>Target word:</div>
            <h2 style={{ margin: '6px 0' }}>
              {roundWord ?? (
                <span style={{ color: '#999' }}>
                  {roundCompleted ? 'Round complete, waiting for next round...' : 'Waiting...'}
                </span>
              )}
            </h2>
            {roundCompleted && lastResult && (
              <div style={{ color: '#222', marginTop: 6 }}>Last result: {lastResult}</div>
            )}
          </div>
          <div>
            <input
              ref={inputRef}
              value={answerInput}
              placeholder={
                roundWord
                  ? 'Type the word and press Enter'
                  : 'Waiting... you can start typing and keep your input'
              }
              onChange={(e) => setAnswerInput((e.target as HTMLInputElement).value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  submitAnswer(answerInput);
                }
              }}
              disabled={waitingForResult}
              style={{ padding: '8px', fontSize: 16, width: '100%', boxSizing: 'border-box' }}
            />
          </div>
        </div>
      </div>

      <div className="log">
        {log.map((l, i) => (
          <div key={i}>{l}</div>
        ))}
        {waitingForResult && (
          <div style={{ marginTop: 8, fontStyle: 'italic' }}>Waiting for opponent...</div>
        )}
      </div>
    </div>
  );
}
