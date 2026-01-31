/**
 * API Client for communicating with the Atoyr backend server
 */

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

export interface StartGameResponse {
  sessionId: string;
  scrambledWord: string;
  scrambledWordDefinition: string;
  token: string;
  remainingSeconds: number;
}

export interface SubmitAnswerResponse {
  correct: boolean;
  score: number;
  attempts: number;
  remainingSeconds: number;
  correctAttemptTimestamps: string[];
  scrambledWord?: string;
  scrambledWordDefinition?: string;
  token?: string;
}

export interface LeaderboardEntry {
  id: string;
  rank: number;
  score: number;
  accuracy: number;
  timestamp: number;
}

export interface LeaderboardResponse {
  entries: LeaderboardEntry[];
  total: number;
}

export class GameAPI {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  async startGame(autoVoice: boolean): Promise<StartGameResponse> {
    const response = await fetch(`${this.baseUrl}/game/start`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ autoVoice }),
    });

    if (!response.ok) {
      throw new Error(`Failed to start game: ${response.statusText}`);
    }

    return response.json();
  }

  async submitAnswer(sessionId: string, token: string, answer: string): Promise<SubmitAnswerResponse> {
    const response = await fetch(`${this.baseUrl}/game/answer`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        sessionId,
        token,
        answer,
      }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || `Failed to submit answer: ${response.statusText}`);
    }

    return response.json();
  }

  subscribeToSSE(sessionId: string, onEvent: (data: Record<string, unknown>) => void, onError: (error: Error) => void): () => void {
    const eventSource = new EventSource(`${this.baseUrl}/game/sse/${sessionId}`);

    const handleEvent = (event: MessageEvent) => {
      console.info('SSE message received:', event.type, event.data);

      try {
        const data = JSON.parse(event.data);
        onEvent({ ...data, type: event.type });
      } catch (err) {
        onError(new Error(`Failed to parse SSE event: ${err}`));
      }
    };

    // Listen for named events from the server
    eventSource.addEventListener('start', handleEvent);
    eventSource.addEventListener('tick', handleEvent);
    eventSource.addEventListener('finish', handleEvent);

    eventSource.onerror = (event) => {
      console.error(event);

      eventSource.close();
      onError(new Error('SSE connection closed'));
    };

    return () => eventSource.close();
  }

  async getLeaderboard(limit: number = 10, page: number = 0): Promise<LeaderboardResponse> {
    const params = new URLSearchParams();
    params.append('limit', limit.toString());
    params.append('page', page.toString());

    const response = await fetch(`${this.baseUrl}/leaderboard?${params.toString()}`);

    if (!response.ok) {
      throw new Error(`Failed to fetch leaderboard: ${response.statusText}`);
    }

    return response.json();
  }
}

export const gameAPI = new GameAPI();
