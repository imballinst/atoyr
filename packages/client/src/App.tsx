import { useEffect, useRef, useState } from 'react';
import HpBar from './components/HpBar';

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
  const [currentRole, setCurrentRole] = useState<string | null>(null);
  const [roundCompleted, setRoundCompleted] = useState(false);
  // Accumulate every turn result for the match (latest first)
  const [lastResults, setLastResults] = useState<string[]>([]);
  const startRef = useRef<number | null>(null);
  const [answerInput, setAnswerInput] = useState('');
  const inputRef = useRef<HTMLInputElement | null>(null);
  const nameInputRef = useRef<HTMLInputElement | null>(null);
  const matchmakeButtonRef = useRef<HTMLButtonElement | null>(null);
  const [nameSubmitted, setNameSubmitted] = useState(false);

  useEffect(() => {
    const name = sessionStorage.getItem('name');
    if (name) setName(name);
  }, []);

  // keep a ref to EventSource so we can close it on unmount or player change
  const esRef = useRef<EventSource | null>(null);
  // helper: pick up to 5 random words from an array in random order
  function pickRandomFive(arr: string[] = []) {
    const copy = arr.slice();
    for (let i = copy.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      const t = copy[i];
      copy[i] = copy[j];
      copy[j] = t;
    }
    return copy.slice(0, 5);
  }
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
        // sanitize incoming players to ensure numeric fields are numbers
        const sanitize = (p: any): Player | null => {
          if (!p) return null;
          return {
            id: p.id,
            name: p.name,
            level: Number(p.level) || 0,
            maxHp: Number(p.maxHp) || 0,
            hp: Number(p.hp) || 0,
            attack: Number(p.attack) || 0,
            defense: Number(p.defense) || 0,
            words: Array.isArray(p.words) ? pickRandomFive(p.words) : [],
          };
        };
        const you: Player | null = sanitize(payload.you);
        const opp: Player | null = sanitize(payload.opponent);
        if (you) setPlayer(you);
        if (opp) setOpponent(opp);
        // set current turn if present
        if (payload.turn !== undefined && payload.turn !== null) setCurrentTurn(payload.turn);
        const oppWords: string[] = (opp && Array.isArray(opp.words) && opp.words) || [];
        setOpponentWords(pickRandomFive(oppWords));
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
        // show potential vs applied damage when present
        const potential = q?.potential ?? q?.damage ?? null;
        const potentialStr = potential !== null ? ` potential=${potential}` : '';
        setLog((l) => [
          `Turn result: ${q?.kind ?? 'unknown'}${potentialStr} applied=${applied ?? 0}`,
          ...(payload.message ? [`Message: ${payload.message}`] : []),
          ...l,
        ]);
        // update local match/opponent state if present
        if (payload.match) {
          const sanitize = (p: any): Player | null => {
            if (!p) return null;
            return {
              id: p.id,
              name: p.name,
              level: Number(p.level) || 0,
              maxHp: Number(p.maxHp) || 0,
              hp: Number(p.hp) || 0,
              attack: Number(p.attack) || 0,
              defense: Number(p.defense) || 0,
              words: Array.isArray(p.words) ? pickRandomFive(p.words) : [],
            };
          };
          const a = sanitize(payload.match.a);
          const b = sanitize(payload.match.b);
          const me = a?.id === forPlayerId ? a : b;
          const opponentObj = a?.id === forPlayerId ? b : a;
          if (me) setPlayer(me);
          if (opponentObj) setOpponent(opponentObj as Player);
          const oppWords: string[] =
            (opponentObj && Array.isArray(opponentObj.words) && opponentObj.words) || [];
          setOpponentWords(pickRandomFive(oppWords));
        }
        // clear waiting state and mark round completed; keep the target visible until next qte_start
        setWaitingForResult(false);
        // clear the input after the turn result so next round starts with an empty field
        setAnswerInput('');
        // clear any fallback waiting timer
        try {
          if (waitingTimerRef.current) {
            window.clearTimeout(waitingTimerRef.current);
            waitingTimerRef.current = null;
          }
        } catch (e) {}
        setRoundCompleted(true);
        // if server sent a personalized message, prefer it; otherwise show generic
        // append the human-friendly result string to lastResults (latest first)
        const resultStr =
          payload.message ?? `${q?.kind ?? 'unknown'}${potentialStr} applied=${applied ?? 0}`;
        setLastResults((s) => [resultStr, ...s]);

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
        // payload: { matchId, turn, word, initiatorId, role }
        setMatchId(payload.matchId);
        setRoundCompleted(false);
        setWaitingForResult(false);
        try {
          if (waitingTimerRef.current) {
            window.clearTimeout(waitingTimerRef.current);
            waitingTimerRef.current = null;
          }
        } catch (e) {}
        // Do not clear the accumulated lastResults here; keep full match history
        setRoundWord(payload.word);
        setRoundIndex((i) => i + 1);
        setCurrentTurn(payload.turn);
        setAnswerInput('');
        if (payload.role) setCurrentRole(String(payload.role));
        startRef.current = performance.now();
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

  // When the window gains focus, focus the appropriate input:
  // - if not in a match, focus the player name input
  // - if in a match, focus the round word input
  useEffect(() => {
    const onFocus = () => {
      try {
        if (!matchId) {
          // if name was submitted, focus the matchmake button so the user can start matchmaking
          if (nameSubmitted) {
            if (matchmakeButtonRef.current) matchmakeButtonRef.current.focus();
          } else {
            // focus player name input when not in match and name not submitted
            if (nameInputRef.current) nameInputRef.current.focus();
          }
        } else {
          // focus the text input for typing the word when in a match
          if (inputRef.current) inputRef.current.focus();
        }
      } catch (e) {}
    };
    window.addEventListener('focus', onFocus);
    return () => window.removeEventListener('focus', onFocus);
  }, [matchId]);

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
      if (data && Array.isArray(data.words) && data.words.length)
        setWords(pickRandomFive(data.words));
      // hide the name input now that the player has been created locally
      setNameSubmitted(true);
      // focus the matchmake button so the user can start matchmaking immediately
      setTimeout(() => {
        try {
          matchmakeButtonRef.current && matchmakeButtonRef.current.focus();
        } catch (e) {}
      }, 50);
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
      setOpponentWords(pickRandomFive(oppWords));
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
      <h1 className="text-2xl font-semibold mb-4">A Toy R — Integrated Prototype</h1>

      {/* Turn indicator */}
      <div className="flex justify-center mb-2">
        <div className="font-bold text-lg px-3 py-1 bg-gray-900 text-white rounded-lg">
          Turn: {currentTurn !== null ? currentTurn : '-'}
          {currentRole ? `: ${currentRole[0].toUpperCase()}${currentRole.slice(1)}` : ''}
        </div>
      </div>

      {/* HP display + words aligned in same block */}
      <div className="flex gap-6 items-start mb-6">
        <div className="flex-1 bg-gray-50 p-3 rounded">
          <div className="font-semibold mb-1">{player?.name ?? player?.id ?? 'You'}</div>
          <HpBar
            value={player?.hp ?? 0}
            max={player?.maxHp ?? 0}
            fromClass="from-green-500"
            toClass="to-lime-400"
            aria-label="Your HP bar"
          />
          <div className="mt-2 text-sm mb-3">
            {player ? `${player.hp} / ${player.maxHp} HP` : 'No player'}
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700">My words (5):</label>
            <ol className="list-decimal list-inside mt-1">
              {words.map((w) => (
                <li key={`me-${w}`} className="py-0.5 text-sm">
                  {w}
                </li>
              ))}
            </ol>
          </div>
        </div>

        <div className="flex-1 bg-gray-50 p-3 rounded">
          <div className="font-semibold mb-1">{opponent?.name ?? opponent?.id ?? 'Opponent'}</div>
          <HpBar
            value={opponent?.hp ?? 0}
            max={opponent?.maxHp ?? 0}
            fromClass="from-red-500"
            toClass="to-rose-400"
            aria-label="Opponent HP bar"
          />
          <div className="mt-2 text-sm mb-3">
            {opponent ? `${opponent.hp} / ${opponent.maxHp} HP` : 'No opponent'}
          </div>
          <div>
            {!matchId && (
              <div className="mb-2">
                <button
                  ref={matchmakeButtonRef}
                  onClick={matchmake}
                  className="px-3 py-2 bg-accent border border-solid rounded"
                >
                  Matchmake (call server)
                </button>
              </div>
            )}
            <label className="block text-sm font-medium text-gray-700">Opponent words:</label>
            <ol className="list-disc list-inside mt-1">
              {opponentWords.map((w) => (
                <li key={`op-${w}`} className="py-0.5 text-sm text-gray-700">
                  {w}
                </li>
              ))}
            </ol>
          </div>
        </div>
      </div>

      {!nameSubmitted && (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            createPlayer();
          }}
          className="mb-4 space-y-2"
        >
          <label className="block text-sm font-medium text-gray-700">Player name:</label>
          <input
            ref={nameInputRef}
            value={name}
            onChange={(e) => {
              setName(e.target.value);
              sessionStorage.setItem('name', e.target.value);
            }}
            className="mt-1 block w-full border rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary"
          />
          <button
            type="submit"
            className="inline-flex items-center px-3 py-2 bg-primary text-white rounded"
          >
            Create Player (local)
          </button>
        </form>
      )}

      <div>
        <div className="mt-3">
          <div className="mb-3">
            <div className="text-sm text-gray-500">Target word:</div>
            <h2 className="mt-1 mb-1 text-xl">
              {roundWord ?? (
                <span className="text-gray-400">
                  {roundCompleted ? 'Round complete, waiting for next round...' : 'Waiting...'}
                </span>
              )}
            </h2>
          </div>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              submitAnswer(answerInput);
            }}
          >
            <input
              ref={inputRef}
              value={answerInput}
              placeholder={
                roundWord
                  ? 'Type the word and press Enter'
                  : 'Waiting... you can start typing and keep your input'
              }
              onChange={(e) => setAnswerInput((e.target as HTMLInputElement).value)}
              disabled={waitingForResult}
              className="w-full px-4 py-3 text-base border rounded focus:outline-none focus:ring-2 focus:ring-primary"
            />
          </form>
        </div>
      </div>

      {lastResults.length > 0 && (
        <div className="text-gray-800 mt-2">
          <div className="font-medium">Last result: {lastResults[0]}</div>
          {lastResults.length > 1 && (
            <ol className="list-decimal list-inside mt-2 text-sm text-gray-700">
              {lastResults.slice(1, 6).map((r, i) => (
                <li key={i}>{r}</li>
              ))}
            </ol>
          )}
        </div>
      )}

      <div className="log" hidden>
        {log.map((l, i) => (
          <div key={i}>{l}</div>
        ))}
        {waitingForResult && <div className="mt-2 italic">Waiting for opponent...</div>}
      </div>
    </div>
  );
}
