// ============================================
// Word Types
// ============================================

export interface WordEntry {
  scrambled: string;
  definition: string;
}

// ============================================
// Game Constants
// ============================================

export const GAME_DURATION_SECONDS = 30;
export const AUTO_VOICE_EXTRA_SECONDS = 5;
export const WRONG_ANSWER_PENALTY_SECONDS = 1;
export const WORD_LENGTH = 5;
export const SCRAMBLE_ATTEMPTS = 10;
export const LEADERBOARD_COOKIE_NAME = 'atoyr_leaderboard';
export const LEADERBOARD_COOKIE_EXPIRY_DAYS = 7;

// ============================================
// Session Types
// ============================================

export type GamePhase = 'idle' | 'playing' | 'finished';

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
}

// ============================================
// Client-side Game State (extends session state)
// ============================================

export interface GameState {
  phase: GamePhase;
  currentWord: WordEntry | null;
  currentWordToken: string | null;
  score: number;
  totalAttempts: number;
  correctAttemptTimestamps: string[];
  remainingSeconds: number;
  usedWords: Set<string>;
  gameResults: GameResult[];
  autoVoice: boolean;
}

// ============================================
// API Request/Response Types
// ============================================

export interface StartGameRequest {
  autoVoice?: boolean;
}

export interface StartGameResponse {
  sessionId: string;
  expiresAt: number;
}

export interface SubmitAnswerRequest {
  sessionId: string;
  token: string; // Word-specific token
  answer: string;
}

export interface SubmitAnswerResponse {
  correct: boolean;
  score: number;
  totalAttempts: number;
  remainingSeconds: number;
  gameOver: boolean;
}

// ============================================
// SSE Event Types
// ============================================

export type SSEEventType = 'session:started' | 'word:new' | 'timer:tick' | 'timer:penalty' | 'game:finished' | 'error';

export interface SSEBaseEvent {
  type: SSEEventType;
  timestamp: number;
}

export interface SessionStartedEvent extends SSEBaseEvent {
  type: 'session:started';
  sessionId: string;
  autoVoice: boolean;
  remainingSeconds: number;
}

export interface WordNewEvent extends SSEBaseEvent {
  type: 'word:new';
  scrambled: string;
  definition: string;
  token: string; // Unique token for this word instance
  wordIndex: number; // 1-based index for display
}

export interface TimerTickEvent extends SSEBaseEvent {
  type: 'timer:tick';
  remainingSeconds: number;
}

export interface TimerPenaltyEvent extends SSEBaseEvent {
  type: 'timer:penalty';
  penaltySeconds: number;
  remainingSeconds: number;
}

export interface GameFinishedEvent extends SSEBaseEvent {
  type: 'game:finished';
  score: number;
  totalAttempts: number;
  accuracy: number;
  resultId: string;
}

export interface SSEErrorEvent extends SSEBaseEvent {
  type: 'error';
  message: string;
  code: string;
}

export type SSEEvent = SessionStartedEvent | WordNewEvent | TimerTickEvent | TimerPenaltyEvent | GameFinishedEvent | SSEErrorEvent;

// ============================================
// Game Result Types
// ============================================

export interface GameResult {
  id: string;
  sessionId?: string; // Optional for backward compatibility with MVP
  timestamp: number;
  score: number;
  totalAttempts: number;
  accuracy: number;
  durationMs?: number; // Optional for backward compatibility with MVP
}

export interface LeaderboardEntry {
  id: string;
  rank: number;
  score: number;
  accuracy: number;
  timestamp: number;
  playerName?: string; // Optional, for future enhancement
}

export interface LeaderboardResponse {
  entries: LeaderboardEntry[];
  totalGames: number;
  lastUpdated: number;
}

// ============================================
// Error Types
// ============================================

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

export const ERROR_MESSAGES: Record<ErrorCode, string> = {
  SESSION_NOT_FOUND: 'Game session not found',
  SESSION_EXPIRED: 'Game session has expired',
  INVALID_TOKEN: 'Invalid word token',
  GAME_NOT_STARTED: 'Game has not been started',
  GAME_ALREADY_FINISHED: 'Game has already finished',
  INVALID_REQUEST: 'Invalid request format',
};
