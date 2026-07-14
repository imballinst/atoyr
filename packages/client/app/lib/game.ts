import { addSeconds, isAfter } from 'date-fns';

export interface WordEntry {
  scrambled: string;
  definition: string;
}

export const GAME_DURATION_SECONDS = 30;

export type GamePhase = 'resuming' | 'idle' | 'playing' | 'finished';

export interface GameSession {
  id: string; // UUID
  createdAt: number; // Unix timestamp
  expiresAt: number; // Unix timestamp
  state: GameSessionState;
  autoVoice: boolean;
}

export interface GameSessionState {
  phase: GamePhase;
  score: number;
  totalAttempts: number;
  remainingSeconds: number;
  usedWords: string[]; // Array for JSON serialization
  currentWordToken: string | null; // Token for current word validation
  correctAttemptTimestamps: string[][];
  autoVoice: boolean;
}

export interface GameState {
  phase: GamePhase;
  currentWord: WordEntry | null;
  currentWordToken: string | null;
  lastWordAnswer: string | null;
  score: number;
  totalAttempts: number;
  correctAttemptTimestamps: string[][];
  remainingSeconds: number;
  usedWords: string[];
  autoVoice: boolean;
}

export type ErrorCode =
  | 'SESSION_NOT_FOUND'
  | 'SESSION_EXPIRED'
  | 'INVALID_TOKEN'
  | 'GAME_NOT_STARTED'
  | 'GAME_ALREADY_FINISHED'
  | 'INVALID_REQUEST';

export interface APIError {
  code: ErrorCode;
  message: string;
  statusCode: number;
}

const LOCAL_STORAGE_GAME_ENDS_AT = 'session-ends-at';

export function getFinalScore(score: number, totalAttempts: number): string {
  let finalScore = `${score}`;

  if (totalAttempts > 0) {
    finalScore += `/${totalAttempts}`;
  }

  return finalScore;
}

export function setGameEndsAt(remainingSeconds: number) {
  localStorage.setItem(LOCAL_STORAGE_GAME_ENDS_AT, addSeconds(new Date(), remainingSeconds).toString());
}
export function hasGameEnded() {
  const ts = localStorage.getItem(LOCAL_STORAGE_GAME_ENDS_AT);
  if (!ts) return true;

  return isAfter(new Date(), new Date(ts));
}
