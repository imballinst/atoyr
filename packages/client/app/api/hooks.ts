/**
 * Server-based game hook that integrates with the backend API
 * Uses SSE for real-time updates and HTTP for answer submissions
 */

import { useQuery } from '@tanstack/react-query';
import { useCallback, useEffect, useRef, useState } from 'react';
import { z } from 'zod';

import { readStoredSettings, writeStoredSettings, type LatestSchema } from '~/lib/settings';

import { GAME_DURATION_SECONDS, setGameEndsAt, type GameState } from '../lib/game';
import { apiQuery, apiResumeGame, apiStartGame, apiSubmitAnswer, apiSubscribeToSSE } from './client';
import type { SessionMode } from './gen';

interface ServerGameSession {
  sessionId: string;
  settings: LatestSchema;
}

const INITIAL_STATE: Omit<GameState, 'settings'> = {
  phase: 'idle',
  currentWord: null,
  currentWordToken: null,
  lastWordAnswer: null,
  score: 0,
  totalAttempts: 0,
  correctAttemptTimestamps: [],
  remainingSeconds: GAME_DURATION_SECONDS,
  usedWords: [],
};
const QUERY_OPTS = { retry: import.meta.env.DEV ? 0 : 3 };

const TickEventSchema = z.object({ type: z.literal('tick'), remainingSeconds: z.number() });
const FinishEventSchema = z.object({ type: z.literal('finish'), lastWordAnswer: z.string() });
const EventSchema = z.union([TickEventSchema, FinishEventSchema]);

export function useGame(shouldContinueGame: boolean, defaultSettings: LatestSchema) {
  const [shouldContinue, setShouldContinue] = useState(shouldContinueGame);
  const [state, setState] = useState<GameState>({
    ...INITIAL_STATE,
    settings: defaultSettings,
    phase: shouldContinueGame ? 'resuming' : 'idle',
  });
  const sessionRef = useRef<ServerGameSession | null>(null);
  const sseUnsubscribeRef = useRef<(() => void) | null>(null);
  const isBeforeUnloadFnRef = useRef<(() => void) | null>(null);
  const isBeforeUnloadRef = useRef(false);

  const startGame = useCallback(async (settings: LatestSchema, action: 'start' | 'resume' = 'start') => {
    try {
      const response = action === 'start' ? await apiStartGame(settings, []) : await apiResumeGame();
      setGameEndsAt(response.remainingSeconds);

      if (isBeforeUnloadFnRef.current) {
        window.removeEventListener('beforeunload', isBeforeUnloadFnRef.current);
      }

      isBeforeUnloadFnRef.current = () => {
        isBeforeUnloadRef.current = true;
      };
      window.addEventListener('beforeunload', isBeforeUnloadFnRef.current);

      sessionRef.current = {
        sessionId: response.sessionId,
        settings,
      };
      setState((prev) => ({
        ...prev,
        phase: 'playing',
        score: 0,
        totalAttempts: 0,
        correctAttemptTimestamps: [],
        remainingSeconds: response.remainingSeconds,
        usedWords: [],
        lastWordAnswer: null,
        currentWord: {
          scrambled: response.scrambledWord,
          definition: response.scrambledWordDefinition,
        },
        currentWordToken: response.token,
      }));

      // Subscribe to SSE events
      if (sseUnsubscribeRef.current) {
        sseUnsubscribeRef.current();
        sseUnsubscribeRef.current = null;
      }

      sseUnsubscribeRef.current = apiSubscribeToSSE(
        async (event) => {
          const { data } = EventSchema.safeParse(event);
          if (!data) return;

          if (data.type === 'tick') {
            setGameEndsAt(data.remainingSeconds);
            setState((prev) => ({
              ...prev,
              remainingSeconds: data.remainingSeconds,
            }));
          } else if (data.type === 'finish') {
            setState((prev) => ({
              ...prev,
              phase: 'finished',
              lastWordAnswer: data.lastWordAnswer,
            }));
            setShouldContinue(false);
          }
        },
        (error) => {
          if (isBeforeUnloadRef.current) return;

          console.error('SSE error:', error);
          setState((prev) => ({
            ...prev,
            phase: 'finished',
          }));
        },
      );
    } catch (err) {
      console.error(`Failed to ${action} game:`, err);

      if (action === 'start') {
        alert(`Failed to ${action} game. Please try again.`);
      }
    }
  }, []);

  const submitAnswer = useCallback(
    async (answer: string, token: string, opts: { onSuccess?: () => void; onError?: () => void }) => {
      if (state.phase !== 'playing' || !sessionRef.current) return;

      const { onSuccess, onError } = opts;

      try {
        const response = await apiSubmitAnswer(token, answer);
        setState((prev) => ({
          ...prev,
          score: response.score,
          correctAttemptTimestamps: response.correctAttemptTimestamps,
          totalAttempts: response.attempts,
          remainingSeconds: response.remainingSeconds,
        }));

        const { scrambledWord: nextScrambledWord, scrambledWordDefinition: nextDefinition, token: nextToken } = response;

        if (response.correct && nextScrambledWord && nextToken) {
          setState((prev) => ({
            ...prev,
            currentWord: {
              scrambled: nextScrambledWord,
              definition: nextDefinition,
            },
            currentWordToken: nextToken,
          }));
          onSuccess?.();
        } else if (response.correct) {
          console.error('Received correct response but missing next word data:', response);
        } else {
          onError?.();
        }
      } catch (err) {
        console.error('Failed to submit answer:', err);
      }
    },
    [state.phase],
  );

  const playAgain = useCallback(() => {
    void startGame(state.settings, 'start');
  }, [state.settings, startGame]);

  const resetGame = useCallback(() => {
    if (sseUnsubscribeRef.current) {
      sseUnsubscribeRef.current();
      sseUnsubscribeRef.current = null;
    }
    sessionRef.current = null;
    isBeforeUnloadFnRef.current = null;

    setState((prev) => ({ ...prev, ...INITIAL_STATE }));
  }, []);

  const updateSettings = useCallback((config: Partial<LatestSchema>) => {
    setState((prev) => ({ ...prev, settings: { ...prev.settings, ...config } }));

    const currentConfig = readStoredSettings();
    writeStoredSettings({ ...currentConfig, ...config });
  }, []);

  useQuery({
    queryKey: ['resumeGame'],
    queryFn: async () => {
      try {
        const response = await apiResumeGame();
        await startGame({ autoVoice: response.autoVoice, mode: response.mode, topic: response.topic }, 'resume');
        return response;
      } catch (err) {
        console.warn('No active session to resume');
        throw err;
      }
    },
    enabled: shouldContinue,
  });

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (sseUnsubscribeRef.current) {
        sseUnsubscribeRef.current();
      }
    };
  }, []);

  return {
    state,
    startGame,
    submitAnswer,
    playAgain,
    updateSettings,
    resetGame,
  };
}

export function useLeaderboard(mode?: SessionMode, page = 1, limit = 10) {
  return apiQuery.useQuery(
    'get',
    '/api/v1/leaderboard',
    {
      params: {
        query: {
          mode,
          page,
          limit,
        },
      },
    },
    QUERY_OPTS,
  );
}

export function useLeaderboardPercentile(mode?: SessionMode) {
  return apiQuery.useQuery(
    'get',
    '/api/v1/leaderboard/percentile',
    {
      params: {
        query: {
          mode,
        },
      },
    },
    QUERY_OPTS,
  );
}
