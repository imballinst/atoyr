/**
 * Server-based game hook that integrates with the backend API
 * Uses SSE for real-time updates and HTTP for answer submissions
 */

import { useCallback, useEffect, useRef, useState } from 'react';
import { gameAPI, SubmitAnswerResponse } from '../api/client';
import { GAME_DURATION_SECONDS, GameResult, GameState, LEADERBOARD_COOKIE_NAME } from '../types/game';

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

export function useServerGame() {
  const [state, setState] = useState<GameState>({
    phase: 'idle',
    currentWord: null,
    currentWordToken: null,
    score: 0,
    totalAttempts: 0,
    remainingSeconds: GAME_DURATION_SECONDS,
    usedWords: new Set(),
    gameResults: [],
    autoVoice: false,
  });

  const sessionRef = useRef<ServerGameSession | null>(null);
  const sseUnsubscribeRef = useRef<(() => void) | null>(null);
  const timerIntervalRef = useRef<any | null>(null);

  const startGame = useCallback((autoVoice: boolean = false) => {
    setState((prev) => ({
      ...prev,
      phase: 'playing',
      score: 0,
      totalAttempts: 0,
      remainingSeconds: autoVoice ? GAME_DURATION_SECONDS + 5 : GAME_DURATION_SECONDS,
      usedWords: new Set(),
      autoVoice,
    }));

    gameAPI
      .startGame(autoVoice)
      .then((response) => {
        sessionRef.current = {
          sessionId: response.sessionId,
          autoVoice,
        };
        setState((prev) => ({
          ...prev,
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
          (event) => {
            const eventType = (event.type as string) || '';

            if (eventType === 'tick') {
              setState((prev) => ({
                ...prev,
                remainingSeconds: event.remainingSeconds as number,
              }));
            } else if (eventType === 'finish') {
              const accuracy = (event.totalAttempts as number) > 0 ? (event.score as number) / (event.totalAttempts as number) : 0;

              const result: GameResult = {
                id: event.resultId as string,
                timestamp: Date.now(),
                score: event.score as number,
                totalAttempts: event.totalAttempts as number,
                accuracy,
              };

              const existingLeaderboard = getCookie(LEADERBOARD_COOKIE_NAME);
              const leaderboard: GameResult[] = existingLeaderboard ? JSON.parse(existingLeaderboard) : [];
              leaderboard.push(result);
              setCookie(LEADERBOARD_COOKIE_NAME, JSON.stringify(leaderboard), 7);

              setState((prev) => ({
                ...prev,
                phase: 'finished',
                score: event.score as number,
                totalAttempts: event.totalAttempts as number,
                gameResults: leaderboard,
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
      })
      .catch((err) => {
        console.error('Failed to start game:', err);
        alert('Failed to start game. Please try again.');
        setState((prev) => ({
          ...prev,
          phase: 'idle',
        }));
      });
  }, []);

  const submitAnswer = useCallback(
    (answer: string, token: string, { onSuccess, onError }: { onSuccess?: () => void; onError?: () => void }) => {
      if (state.phase !== 'playing' || !sessionRef.current) return;

      const session = sessionRef.current;

      gameAPI
        .submitAnswer(session.sessionId, token, answer)
        .then((response: SubmitAnswerResponse) => {
          setState((prev) => ({
            ...prev,
            score: response.score,
            totalAttempts: response.attempts,
            remainingSeconds: response.remainingSeconds,
          }));

          const { scrambledWord, scrambledWordDefinition, token } = response;

          if (response.correct && scrambledWord && scrambledWordDefinition && token) {
            setState((prev) => ({
              ...prev,
              currentWord: {
                scrambled: scrambledWord,
                definition: scrambledWordDefinition,
              },
              currentWordToken: token,
            }));
            onSuccess?.();
          } else {
            onError?.();
          }
        })
        .catch((err) => {
          console.error('Failed to submit answer:', err);
          // Allow user to retry
        });
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
      phase: 'idle',
      currentWord: null,
      scrambled: null,
      currentWordToken: null,
      score: 0,
      totalAttempts: 0,
      remainingSeconds: GAME_DURATION_SECONDS,
      usedWords: new Set(),
      gameResults: prev.gameResults,
      autoVoice: false,
    }));
  }, []);

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
