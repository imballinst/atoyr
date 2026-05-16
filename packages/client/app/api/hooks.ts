/**
 * Server-based game hook that integrates with the backend API
 * Uses SSE for real-time updates and HTTP for answer submissions
 */

import { useQuery } from '@tanstack/react-query';
import { isAxiosError } from 'axios';
import { useCallback, useEffect, useRef, useState } from 'react';
import { GAME_DURATION_SECONDS, type GameState } from '../lib/game';
import { apiQuery, apiResumeGame, apiStartGame, apiSubmitAnswer, apiSubscribeToSSE } from './client';

interface ServerGameSession {
  sessionId: string;
  autoVoice: boolean;
}

const INITIAL_STATE: GameState = {
  phase: 'idle',
  currentWord: null,
  currentWordToken: null,
  score: 0,
  totalAttempts: 0,
  correctAttemptTimestamps: [],
  remainingSeconds: GAME_DURATION_SECONDS,
  usedWords: [],
  autoVoice: false,
};

export function useGame() {
  const [state, setState] = useState(INITIAL_STATE);

  const sessionRef = useRef<ServerGameSession | null>(null);
  const sseUnsubscribeRef = useRef<(() => void) | null>(null);
  const timerIntervalRef = useRef<any | null>(null);

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
      sseUnsubscribeRef.current = apiSubscribeToSSE(
        response.sessionId,
        async (event) => {
          const eventType = (event.type as string) || '';

          if (eventType === 'tick') {
            setState((prev) => ({
              ...prev,
              remainingSeconds: event.remainingSeconds as number,
            }));
          } else if (eventType === 'finish') {
            setState((prev) => ({
              ...prev,
              phase: 'finished',
              correctAttemptTimestamps: event.correctAttemptTimestamps as string[][],
              score: event.score as number,
              totalAttempts: event.totalAttempts as number,
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
    if (timerIntervalRef.current) {
      clearInterval(timerIntervalRef.current);
      timerIntervalRef.current = null;
    }
    sessionRef.current = null;

    setState(INITIAL_STATE);
  }, []);

  useQuery({
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
  });

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (sseUnsubscribeRef.current) {
        sseUnsubscribeRef.current();
      }
      if (timerIntervalRef.current) {
        clearInterval(timerIntervalRef.current);
      }
    };
  }, []);

  return {
    state,
    startGame,
    submitAnswer,
    resetGame,
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
