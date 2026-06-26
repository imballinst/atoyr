export type SSEEventType = 'session:started' | 'word:new' | 'timer:tick' | 'timer:penalty' | 'game:finished' | 'error';

export interface SSEBaseEvent {
  type: SSEEventType;
  timestamp: number;
}

export interface SessionStartedEvent extends SSEBaseEvent {
  type: 'session:started';
  sessionId: string;
  autoVoice: boolean;
  durationSeconds: number;
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
  durationSeconds: number;
}

export interface TimerPenaltyEvent extends SSEBaseEvent {
  type: 'timer:penalty';
  penaltySeconds: number;
  durationSeconds: number;
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
