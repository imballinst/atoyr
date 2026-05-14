import createFetchClient from 'openapi-fetch';
import createClient from 'openapi-react-query';
import type { paths } from './gen';

const API_BASE_URL = import.meta.env.VITE_API_URL || '';
const CLIENT = createFetchClient<paths>({
  baseUrl: API_BASE_URL,
});

export const client = CLIENT;
export const query = createClient(CLIENT);

export class GameAPI {
  async resumeGame() {
    const response = await CLIENT.POST('/api/v1/game/continue');
    if (!response.data) {
      throw new Error(response.error);
    }

    return response.data;
  }

  async startGame(autoVoice: boolean, itemsUsed: string[]) {
    const response = await CLIENT.POST('/api/v1/game/start', {
      body: {
        autoVoice,
        itemsUsed,
      },
    });
    if (!response.data) {
      throw new Error(response.error);
    }

    return response.data;
  }

  async submitAnswer(sessionId: string, token: string, answer: string) {
    const response = await CLIENT.POST('/api/v1/game/submit', {
      body: {
        sessionId,
        token,
        answer,
      },
    });
    if (!response.data) {
      throw new Error(response.error);
    }

    return response.data;
  }

  subscribeToSSE(sessionId: string, onEvent: (data: Record<string, unknown>) => void, onError: (error: Error) => void): () => void {
    const eventSourceURL: keyof paths = '/api/v1/game/sse/{sessionId}';
    const eventSource = new EventSource(eventSourceURL.replace('{sessionId}', sessionId));

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

  async getLeaderboard(limit: number = 10, page: number = 0) {
    const response = await CLIENT.GET('/api/v1/leaderboard', {
      params: {
        query: {
          page,
          limit,
        },
      },
    });
    if (!response.data) {
      throw new Error(response.error);
    }

    return response.data;
  }

  async getInventory() {
    const response = await CLIENT.GET('/api/v1/items/me');
    if (!response.data) {
      throw new Error(response.error);
    }

    return response.data;
  }
}

export const gameAPI = new GameAPI();
