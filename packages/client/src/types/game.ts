export interface WordEntry {
  word: string;
  definition: string;
}

export interface GameResult {
  id: string;
  timestamp: number;
  score: number;
  totalAttempts: number;
  accuracy: number;
}

export type GamePhase = 'idle' | 'playing' | 'finished';

export interface GameState {
  phase: GamePhase;
  currentWord: WordEntry | null;
  scrambled: string | null;
  score: number;
  totalAttempts: number;
  remainingSeconds: number;
  usedWords: Set<string>;
  gameResults: GameResult[];
}

export const GAME_DURATION_SECONDS = 30;
export const WRONG_ANSWER_PENALTY_SECONDS = 1;
export const LEADERBOARD_COOKIE_NAME = 'atoyr_leaderboard';
export const LEADERBOARD_COOKIE_EXPIRY_DAYS = 7;
