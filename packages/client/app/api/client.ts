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

const EVENT_SOURCE_URL: keyof paths = '/api/v1/game/sse';
const HEARTBEAT_TIMEOUT = 5000;
const HEARTBEAT_INTERVAL = 1000;

export function apiSubscribeToSSE(onEvent: (data: Record<string, unknown>) => void, onError: (error: Error) => void): () => void {
  let eventSource: EventSource;
  let heartbeatInterval: ReturnType<typeof setInterval>;

  const cleanup = () => {
    clearInterval(heartbeatInterval);
    eventSource?.close();
  };

  const reconnect = () => {
    clearInterval(heartbeatInterval);
    setTimeout(() => {
      const result = initSSE(onEvent, reconnect, onError);
      eventSource = result.eventSource;
      heartbeatInterval = result.heartbeatInterval;
    }, 1000);
  };

  const result = initSSE(onEvent, reconnect, onError);
  eventSource = result.eventSource;
  heartbeatInterval = result.heartbeatInterval;

  return cleanup;
}

function initSSE(onEvent: (data: Record<string, unknown>) => void, onReconnect: () => void, onError: (error: Error) => void) {
  const eventSource = new EventSource(EVENT_SOURCE_URL);
  let lastMessageReceived: string | undefined;
  let lastEventTime = Date.now();

  const heartbeatInterval = setInterval(() => {
    if (Date.now() - lastEventTime > HEARTBEAT_TIMEOUT) {
      clearInterval(heartbeatInterval);
      eventSource.close();
      if (lastMessageReceived !== 'finish') {
        onReconnect();
      }
    }
  }, HEARTBEAT_INTERVAL);

  const handleEvent = (event: MessageEvent) => {
    if (import.meta.env.DEV) {
      console.info('SSE message received:', event.type, event.data);
    }

    lastMessageReceived = event.type;
    lastEventTime = Date.now();

    try {
      const data = JSON.parse(event.data);
      onEvent({ ...data, type: event.type });
    } catch (err) {
      onError(new Error(`Failed to parse SSE event: ${err}`));
    }
  };

  eventSource.addEventListener('start', handleEvent);
  eventSource.addEventListener('tick', handleEvent);
  eventSource.addEventListener('finish', handleEvent);

  eventSource.onerror = () => {
    clearInterval(heartbeatInterval);
    eventSource.close();

    if (lastMessageReceived !== 'finish') {
      // Reconnect only if the last state isn't finish.
      onReconnect();
    }
  };

  return { eventSource, heartbeatInterval };
}
