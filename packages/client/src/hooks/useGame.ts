import { useCallback, useEffect, useRef, useState } from 'react';
import wordData from '../data/words.json';
import {
  GAME_DURATION_SECONDS,
  GameResult,
  GameState,
  LEADERBOARD_COOKIE_NAME,
  WordEntry,
  WRONG_ANSWER_PENALTY_SECONDS,
} from '../types/game';
import { scrambleWord } from '../utils/scramble';
import { validateAnswer } from '../utils/validation';

function getRandomWord(usedWords: Set<string>): WordEntry | null {
  const available = (wordData as WordEntry[]).filter((w) => !usedWords.has(w.word));
  if (available.length === 0) return null;
  return available[Math.floor(Math.random() * available.length)];
}

function generateId(): string {
  return Math.random().toString(36).substring(2, 9);
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

export function useGame() {
  const [state, setState] = useState<GameState>({
    phase: 'idle',
    currentWord: null,
    scrambled: null,
    score: 0,
    totalAttempts: 0,
    remainingSeconds: GAME_DURATION_SECONDS,
    usedWords: new Set(),
    gameResults: [],
  });

  const timerIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const startGame = useCallback(() => {
    const initialWord = getRandomWord(new Set());
    if (!initialWord) {
      alert('No words available');
      return;
    }

    const scrambled = scrambleWord(initialWord.word);
    if (!scrambled) {
      // Skip and try another word
      setState((prev) => ({
        ...prev,
        usedWords: new Set([...prev.usedWords, initialWord.word]),
      }));
      return;
    }

    setState((prev) => ({
      ...prev,
      phase: 'playing',
      currentWord: initialWord,
      scrambled,
      score: 0,
      totalAttempts: 0,
      remainingSeconds: GAME_DURATION_SECONDS,
      usedWords: new Set([initialWord.word]),
    }));
  }, []);

  useEffect(() => {
    if (state.phase !== 'playing') return;

    timerIntervalRef.current = setInterval(() => {
      setState((prev) => {
        const next = prev.remainingSeconds - 1;
        if (next <= 0) {
          return {
            ...prev,
            phase: 'finished',
            remainingSeconds: 0,
          };
        }
        return { ...prev, remainingSeconds: next };
      });
    }, 1000);

    return () => {
      if (timerIntervalRef.current) {
        clearInterval(timerIntervalRef.current);
      }
    };
  }, [state.phase]);

  useEffect(() => {
    if (state.phase === 'finished' && state.currentWord) {
      const accuracy = state.totalAttempts > 0 ? state.score / state.totalAttempts : 0;
      const result: GameResult = {
        id: generateId(),
        timestamp: Date.now(),
        score: state.score,
        totalAttempts: state.totalAttempts,
        accuracy,
      };

      const existingLeaderboard = getCookie(LEADERBOARD_COOKIE_NAME);
      const leaderboard: GameResult[] = existingLeaderboard ? JSON.parse(existingLeaderboard) : [];
      leaderboard.push(result);
      setCookie(LEADERBOARD_COOKIE_NAME, JSON.stringify(leaderboard), 7);

      setState((prev) => ({
        ...prev,
        gameResults: leaderboard,
      }));
    }
  }, [state.phase]);

  const submitAnswer = useCallback(
    (answer: string) => {
      if (state.phase !== 'playing' || !state.currentWord) return;

      const isCorrect = validateAnswer(answer, state.currentWord.word);

      if (isCorrect) {
        const nextWord = getRandomWord(state.usedWords);
        if (!nextWord) {
          setState((prev) => ({
            ...prev,
            phase: 'finished',
            score: prev.score + 1,
            totalAttempts: prev.totalAttempts + 1,
          }));
          return;
        }

        let scrambled = scrambleWord(nextWord.word);
        while (!scrambled) {
          const anotherWord = getRandomWord(new Set([...state.usedWords, nextWord.word]));
          if (!anotherWord) {
            setState((prev) => ({
              ...prev,
              phase: 'finished',
              score: prev.score + 1,
              totalAttempts: prev.totalAttempts + 1,
            }));
            return;
          }
          scrambled = scrambleWord(anotherWord.word);
          if (scrambled) {
            setState((prev) => ({
              ...prev,
              currentWord: anotherWord,
              scrambled,
              score: prev.score + 1,
              totalAttempts: prev.totalAttempts + 1,
              usedWords: new Set([...prev.usedWords, anotherWord.word]),
            }));
            return;
          }
        }

        setState((prev) => ({
          ...prev,
          currentWord: nextWord,
          scrambled,
          score: prev.score + 1,
          totalAttempts: prev.totalAttempts + 1,
          usedWords: new Set([...prev.usedWords, nextWord.word]),
        }));
      } else {
        console.debug(`Incorrect answer: ${answer}. Expected: ${state.currentWord.word}.`);

        const penalty = Math.min(WRONG_ANSWER_PENALTY_SECONDS, state.remainingSeconds);
        setState((prev) => ({
          ...prev,
          totalAttempts: prev.totalAttempts + 1,
          remainingSeconds: Math.max(0, prev.remainingSeconds - penalty),
        }));
      }
    },
    [state],
  );

  const resetGame = useCallback(() => {
    setState({
      phase: 'idle',
      currentWord: null,
      scrambled: null,
      score: 0,
      totalAttempts: 0,
      remainingSeconds: GAME_DURATION_SECONDS,
      usedWords: new Set(),
      gameResults: state.gameResults,
    });
  }, [state.gameResults]);

  return {
    state,
    startGame,
    submitAnswer,
    resetGame,
  };
}
