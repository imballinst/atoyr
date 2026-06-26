import createFetchClient from 'openapi-fetch';
import createClient from 'openapi-react-query';

import type { paths } from './gen';

const API_BASE_URL = import.meta.env.VITE_API_URL || '';

export const apiClient = createFetchClient<paths>({
  baseUrl: API_BASE_URL,
});
export const apiQuery = createClient(apiClient);

export async function apiResumeGame() {
  const response = await apiClient.POST('/api/v1/game/continue');
  if (!response.data) {
    throw new Error(response.error);
  }

  return response.data;
}

export async function apiStartGame(autoVoice: boolean, itemsUsed: string[]) {
  const response = await apiClient.POST('/api/v1/game/start', {
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

export async function apiSubmitAnswer(token: string, answer: string) {
  const response = await apiClient.POST('/api/v1/game/submit', {
    body: {
      token,
      answer,
    },
  });
  if (!response.data) {
    throw new Error(response.error);
  }

  return response.data;
}

export async function apiGetPercentile() {
  const response = await apiClient.GET('/api/v1/leaderboard/percentile');
  if (!response.data) {
    throw new Error(response.error);
  }

  return response.data;
}

export function apiSubscribeToSSE(onEvent: (data: Record<string, unknown>) => void, onError: (error: Error) => void): () => void {
  const eventSourceURL: keyof paths = '/api/v1/game/sse';
  const eventSource = new EventSource(eventSourceURL);

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
