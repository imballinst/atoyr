import * as Sentry from '@sentry/browser';
import createFetchClient, { type Middleware } from 'openapi-fetch';
import createClient from 'openapi-react-query';

import type { paths } from './gen';

const API_BASE_URL = import.meta.env.VITE_API_URL || '';

class ApiError extends Error {
  responseBody: string;

  constructor(message: string, responseBody: Record<string, any>) {
    super(message);
    this.name = 'ApiError';
    this.responseBody = JSON.stringify(responseBody);
  }
}

export const apiClient = createFetchClient<paths>({
  baseUrl: API_BASE_URL,
});
const throwOnErrorMiddleware: Middleware = {
  async onResponse({ request, response }) {
    if (!response.ok && response.status) {
      const method = request.method.toUpperCase();
      const url = new URL(request.url);
      const path = url.pathname;
      let body: Record<string, any> = {};

      try {
        body = await response.json();
      } catch {
        // No-op.
      }

      const error = new ApiError(`Error when fetching ${method} ${path}: ${response.status} ${response.statusText}`, body);
      console.error(error);

      Sentry.captureException(error);
      throw error;
    }
  },
};
apiClient.use(throwOnErrorMiddleware);

export const apiQuery = createClient(apiClient);

export async function apiResumeGame() {
  const response = await apiClient.POST('/api/v1/game/continue');
  return response.data!;
}

export async function apiStartGame(autoVoice: boolean, itemsUsed: string[]) {
  const response = await apiClient.POST('/api/v1/game/start', {
    body: {
      autoVoice,
      itemsUsed,
    },
  });
  return response.data!;
}

export async function apiSubmitAnswer(token: string, answer: string) {
  const response = await apiClient.POST('/api/v1/game/submit', {
    body: {
      token,
      answer,
    },
  });
  return response.data!;
}

export async function apiGetPercentile() {
  const response = await apiClient.GET('/api/v1/leaderboard/percentile');
  return response.data!;
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
      Sentry.captureException(err);
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
