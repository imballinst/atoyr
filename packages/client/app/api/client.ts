import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

export interface StartGameResponse {
  sessionId: string;
  scrambledWord: string;
  scrambledWordDefinition: string;
  token: string;
  remainingSeconds: number;
  autoVoice: boolean;
}

export interface SubmitAnswerResponse {
  correct: boolean;
  score: number;
  attempts: number;
  remainingSeconds: number;
  correctAttemptTimestamps: string[][];
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
  user?: {
    id: string;
    username: string;
  };
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

  async resumeGame(): Promise<StartGameResponse> {
    const response = await axios.post<StartGameResponse>(`${this.baseUrl}/game/continue`);
    return response.data;
  }

  async startGame(autoVoice: boolean, itemsUsed: string[]): Promise<StartGameResponse> {
    const response = await axios.post<StartGameResponse>(`${this.baseUrl}/game/start`, {
      autoVoice,
      itemsUsed,
    });
    return response.data;
  }

  async submitAnswer(sessionId: string, token: string, answer: string): Promise<SubmitAnswerResponse> {
    const response = await axios.post<SubmitAnswerResponse>(`${this.baseUrl}/game/answer`, {
      sessionId,
      token,
      answer,
    });
    return response.data;
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

    const response = await axios.get<LeaderboardResponse>(`${this.baseUrl}/leaderboard`, { params });
    return response.data;
  }

  async getInventory(
    limit: number = 10,
    page: number = 0,
  ): Promise<{ items: { id: string; name: string; description: string }[]; total: number }> {
    const params = new URLSearchParams();
    params.append('limit', limit.toString());
    params.append('page', page.toString());

    const response = await axios.get<{ items: { id: string; name: string; description: string }[]; total: number }>(
      `${this.baseUrl}/inventory`,
      { params },
    );
    return response.data;
  }
}

export const gameAPI = new GameAPI();
