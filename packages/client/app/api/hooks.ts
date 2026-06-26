/**
 * Server-based game hook that integrates with the backend API
 * Uses SSE for real-time updates and HTTP for answer submissions
 */

import { useQuery } from '@tanstack/react-query';
import { isAxiosError } from 'axios';
import { addSeconds, isAfter } from 'date-fns';
import { useCallback, useEffect, useRef, useState } from 'react';
import { z } from 'zod';

import { GAME_DURATION_SECONDS, setGameEndsAt, type GameState } from '../lib/game';
import { apiQuery, apiResumeGame, apiStartGame, apiSubmitAnswer, apiSubscribeToSSE } from './client';

interface ServerGameSession {
  sessionId: string;
  autoVoice: boolean;
}

const INITIAL_STATE: GameState = {
  phase: 'idle',
  currentWord: null,
  currentWordToken: null,
  lastWordAnswer: null,
  score: 0,
  totalAttempts: 0,
  correctAttemptTimestamps: [],
  remainingSeconds: GAME_DURATION_SECONDS,
  usedWords: [],
  autoVoice: false,
};

const TickEventSchema = z.object({ type: z.literal('tick'), remainingSeconds: z.number() });
const FinishEventSchema = z.object({ type: z.literal('finish'), lastWordAnswer: z.string() });
const EventSchema = z.union([TickEventSchema, FinishEventSchema]);

export function useGame(shouldContinueGame: boolean) {
  const [state, setState] = useState(INITIAL_STATE);

  const sessionRef = useRef<ServerGameSession | null>(null);
  const sseUnsubscribeRef = useRef<(() => void) | null>(null);

  const startGame = useCallback(async (autoVoice: boolean = false, action: 'start' | 'resume' = 'start') => {
    setState((prev) => ({
      ...prev,
      score: 0,
      totalAttempts: 0,
      remainingSeconds: autoVoice ? GAME_DURATION_SECONDS + 5 : GAME_DURATION_SECONDS,
      usedWords: [],
      autoVoice,
    }));

    try {
      const response = action === 'start' ? await apiStartGame(autoVoice, []) : await apiResumeGame();
      sessionRef.current = {
        sessionId: response.sessionId,
        autoVoice,
      };
      setState((prev) => ({
        ...prev,
        phase: 'playing',
        remainingSeconds: response.remainingSeconds,
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
        response.sessionId,
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
          }
        },
        (error) => {
          console.error('SSE error:', error);
          setState((prev) => ({
            ...prev,
            phase: 'finished',
          }));
        },
      );
    } catch (err) {
      if (isAxiosError(err) && err.response && err.response.status >= 400 && err.response.status < 500) {
        return;
      }

      console.error(`Failed to ${action} game:`, err);

      if (action === 'start') {
        alert(`Failed to ${action} game. Please try again.`);
      }
    }
  }, []);

  const submitAnswer = useCallback(
    async (answer: string, token: string, opts: { onSuccess?: () => void; onError?: () => void }) => {
      if (state.phase !== 'playing' || !sessionRef.current) return;

      const session = sessionRef.current;
      const { onSuccess, onError } = opts;

      try {
        const response = await apiSubmitAnswer(session.sessionId, token, answer);
        setState((prev) => ({
          ...prev,
          score: response.score,
          correctAttemptTimestamps: response.correctAttemptTimestamps,
          totalAttempts: response.attempts,
          remainingSeconds: response.remainingSeconds,
        }));

        const { scrambledWord: nextScrambledWord, scrambledWordDefinition: nextDefinition, token: nextToken } = response;

        if (response.correct && nextScrambledWord && nextDefinition && nextToken) {
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

  const resetGame = useCallback(() => {
    if (sseUnsubscribeRef.current) {
      sseUnsubscribeRef.current();
      sseUnsubscribeRef.current = null;
    }
    sessionRef.current = null;

    setState(INITIAL_STATE);
  }, []);

  const resumeGameQuery = useQuery({
    queryKey: ['resumeGame'],
    queryFn: async () => {
      try {
        const response = await apiResumeGame();
        startGame(response.autoVoice, 'resume');
        return response;
      } catch (err) {
        console.warn('No active session to resume');
        throw err;
      }
    },
    enabled: shouldContinueGame,
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
    resetGame,
    resumeGameQuery,
  };
}

export function useLeaderboard(page = 1, limit = 10) {
  return apiQuery.useQuery('get', '/api/v1/leaderboard', {
    params: {
      query: {
        page,
        limit,
      },
    },
  });
}

export function useInventory() {
  return apiQuery.useQuery('get', '/api/v1/items/me');
}
