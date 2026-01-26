import { SSEEvent } from '@atoyr/shared';
import { INestApplication } from '@nestjs/common';
import { NestFactory } from '@nestjs/core';
import EventSource from 'eventsource';
import { AppModule } from '../src/app.module';
import { HttpExceptionFilter } from '../src/common/filters/http-exception.filter';

export async function createTestApp(): Promise<INestApplication> {
  const app = await NestFactory.create(AppModule);
  app.enableCors({
    origin: '*',
    credentials: true,
  });
  app.useGlobalFilters(new HttpExceptionFilter());
  await app.init();
  return app;
}

export function createSSEClient(url: string): EventSource {
  return new EventSource(url);
}

export function waitForSSEEvent<T extends SSEEvent>(es: EventSource, eventType: string, timeout = 5000): Promise<T> {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error(`Timeout waiting for event: ${eventType}`));
    }, timeout);

    es.addEventListener(eventType, (event: MessageEvent) => {
      clearTimeout(timer);
      resolve(JSON.parse(event.data));
    });
  });
}

export function collectSSEEvents(es: EventSource, eventTypes: string[], timeout = 10000): Promise<Record<string, SSEEvent[]>> {
  const events: Record<string, SSEEvent[]> = {};
  eventTypes.forEach((type) => (events[type] = []));

  return new Promise((resolve) => {
    const timer = setTimeout(() => resolve(events), timeout);

    eventTypes.forEach((type) => {
      es.addEventListener(type, (event: MessageEvent) => {
        events[type].push(JSON.parse(event.data));
      });
    });

    es.addEventListener('game:finished', () => {
      clearTimeout(timer);
      setTimeout(() => resolve(events), 100);
    });
  });
}

export async function startGameSession(app: INestApplication, autoVoice = false) {
  const request = (await import('supertest')).default;
  const response = await request(app.getHttpServer()).post('/api/game/start').send({ autoVoice }).expect(201);

  return { sessionId: response.body.sessionId };
}

export async function startGameSessionWithCustomTime(app: INestApplication, seconds: number) {
  const request = (await import('supertest')).default;
  const response = await request(app.getHttpServer())
    .post('/api/game/start/test')
    .send({ autoVoice: false, durationSeconds: seconds })
    .expect(201);

  return { sessionId: response.body.sessionId };
}
