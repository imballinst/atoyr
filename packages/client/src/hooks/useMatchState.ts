import { GameEventPayload, MatchState, SSEEventType } from '@atoyr/shared';
import { useEffect, useReducer, useRef, useState } from 'react';

interface TimerInfo {
  startTimestamp: string;
  endTimestamp: string;
  name: string;
}

const NEXT_TURN_TIMER_NAME = 'Next turn';
const TURN_END_TIMER_NAME = 'Turn end';

export function useMatchState(playerId: string | undefined) {
  const [timerInfo, setTimerInfo] = useState<TimerInfo | null>(null);
  const [matchState, setMatchState] = useReducer(matchStateReducer, null);
  const esRef = useRef<EventSource | null>(null);

  function initSSE(matchId: string, playerId: string) {
    if (esRef.current) {
      try {
        esRef.current.close();
      } catch (e) {}

      esRef.current = null;
    }

    const url = `/api/events/${encodeURIComponent(playerId)}`;
    const es = new EventSource(url);
    esRef.current = es;

    es.onopen = () => {
      setMatchState((prev) => ({
        ...prev,
        rawLogs: [`SSE connection opened for ${playerId}`, ...prev.rawLogs],
      }));
    };

    es.addEventListener(SSEEventType.MESSAGE, (ev: MessageEvent) => {
      setMatchState((prev) => ({
        ...prev,
        rawLogs: [`SSE message: ${ev.data}`, ...prev.rawLogs],
      }));
    });

    es.addEventListener(SSEEventType.MATCH_READY, (ev: MessageEvent) => {
      try {
        const payload = JSON.parse((ev as MessageEvent).data);
        const parsed = MatchState.parse(payload);

        setMatchState((prev) => ({
          ...prev,
          ...parsed,
        }));
        setMatchState((prev) => ({
          ...prev,
          rawLogs: [`Match ready: opponent=${parsed.opponent?.id ?? 'unknown'}`, ...prev.rawLogs],
        }));
      } catch (e) {
        setMatchState((prev) => ({
          ...prev,
          rawLogs: [`match_ready parse error: ${(e as Error).message}`, ...prev.rawLogs],
        }));
      }
    });

    es.addEventListener(SSEEventType.TURN_RESULT, async (ev: any) => {
      try {
        const payload = JSON.parse((ev as MessageEvent).data);
        const parsed = GameEventPayload.parse(payload);
        if (parsed.kind !== 'turn_result') return;

        setMatchState((prev) => ({
          ...prev,
          ...parsed.nextTurnState,
          rawLogs: [`Turn result: ${parsed.result}. Damage dealt: ${parsed.damage}`, ...prev.rawLogs],
        }));

        const now = new Date().toISOString();
        setTimerInfo({
          startTimestamp: now,
          endTimestamp: parsed.nextTurnTimestamp,
          name: NEXT_TURN_TIMER_NAME,
        });
        const delay = new Date(parsed.nextTurnTimestamp).getTime() - new Date(now).getTime();

        setTimeout(() => {
          setTimerInfo((prev) => {
            if (prev?.name !== NEXT_TURN_TIMER_NAME) return prev;
            return null;
          });
        }, delay);

        await fetch(`/api/events/${encodeURIComponent(playerId)}/ack`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ matchId, turn: payload.turn }),
        });
      } catch (e) {
        setMatchState((prev) => ({
          ...prev,
          rawLogs: [`${SSEEventType.TURN_RESULT} error: ${(e as Error).message}`, ...prev.rawLogs],
        }));
      }
    });

    es.addEventListener(SSEEventType.QTE_START, (ev: any) => {
      try {
        const payload = JSON.parse((ev as MessageEvent).data);
        const parsed = GameEventPayload.parse(payload);
        if (parsed.kind !== 'turn_start') return;

        const now = new Date().toISOString();
        setTimerInfo({
          startTimestamp: now,
          endTimestamp: parsed.turnEndTimestamp,
          name: TURN_END_TIMER_NAME,
        });
        const delay = new Date(parsed.turnEndTimestamp).getTime() - new Date(now).getTime();

        setTimeout(() => {
          setTimerInfo((prev) => {
            if (prev?.name !== TURN_END_TIMER_NAME) return prev;
            return null;
          });
        }, delay);
      } catch (e) {
        setMatchState((prev) => ({
          ...prev,
          rawLogs: [`${SSEEventType.QTE_START} error: ${(e as Error).message}`, ...prev.rawLogs],
        }));
      }
    });

    es.onerror = (err) => {
      console.error(err);
      setMatchState((prev) => ({
        ...prev,
        rawLogs: [`SSE error: ${String(err)}`, ...prev.rawLogs],
      }));
    };

    return es;
  }

  useEffect(() => {
    let es: EventSource;
    if (playerId && matchState?.id) {
      es = initSSE(matchState?.id, playerId);
    }

    return () => {
      if (es) {
        try {
          es.close();
        } catch (e) {}
      }
      esRef.current = null;
    };
  }, [playerId, matchState?.id]);

  return { timerInfo, matchState, setMatchState };
}

function matchStateReducer(
  state: MatchState | null,
  action: Partial<MatchState> | ((prevState: MatchState) => MatchState),
): MatchState | null {
  if (state === null) return state;

  if (typeof action === 'function') {
    return action(state);
  }

  return {
    ...state,
    ...action,
  };
}
