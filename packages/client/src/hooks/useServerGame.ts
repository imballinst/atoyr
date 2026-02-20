/**
 * Server-based game hook that integrates with the backend API
 * Uses SSE for real-time updates and HTTP for answer submissions
 */

import { useCallback, useEffect, useRef, useState } from 'react';
import { gameAPI } from '../api/client';
import { GAME_DURATION_SECONDS, GameState } from '../types/game';

interface ServerGameSession {
  sessionId: string;
  autoVoice: boolean;
}

function setCookie(name: string, value: string, days: number): void {
  const date = new Date();
  date.setTime(date.getTime() + days * 24 * 60 * 60 * 1000);
  const expires = `expires=${date.toUTCString()}`;
  document.cookie = `${name}=${value};${expires};path=/`;
}

function getCookie(name: string): string | null {
  const nameEQ = `${name}=`;
  const cookies = document.cookie.split(';');
  for (const cookie of cookies) {
    const trimmed = cookie.trim();
    if (trimmed.startsWith(nameEQ)) {
      return trimmed.substring(nameEQ.length);
    }
  }
  return null;
}

const INITIAL_STATE: GameState = {
  phase: 'idle',
  currentWord: null,
  currentWordToken: null,
  score: 0,
  totalAttempts: 0,
  correctAttemptTimestamps: [],
  remainingSeconds: GAME_DURATION_SECONDS,
  usedWords: new Set(),
  gameResults: [],
  autoVoice: false,
};

export function useServerGame() {
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
      usedWords: new Set(),
      autoVoice,
    }));

    try {
      const response = action === 'start' ? await gameAPI.startGame(autoVoice, []) : await gameAPI.resumeGame();
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
      sseUnsubscribeRef.current = gameAPI.subscribeToSSE(
        response.sessionId,
        async (event) => {
          const eventType = (event.type as string) || '';

          if (eventType === 'tick') {
            setState((prev) => ({
              ...prev,
              remainingSeconds: event.remainingSeconds as number,
            }));
          } else if (eventType === 'finish') {
            const leaderboardResponse = await gameAPI.getLeaderboard();

            setState((prev) => ({
              ...prev,
              phase: 'finished',
              correctAttemptTimestamps: event.correctAttemptTimestamps as string[][],
              score: event.score as number,
              totalAttempts: event.totalAttempts as number,
              gameResults: leaderboardResponse.entries,
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
        const response = await gameAPI.submitAnswer(session.sessionId, token, answer);
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

    setState((prev) => ({
      ...INITIAL_STATE,
      gameResults: prev.gameResults,
    }));
  }, []);

  useEffect(() => {
    startGame(undefined, 'resume');
  }, [startGame]);

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
