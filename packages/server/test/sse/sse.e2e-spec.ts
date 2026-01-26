import {
  AUTO_VOICE_EXTRA_SECONDS,
  GAME_DURATION_SECONDS,
  SessionStartedEvent,
  TimerPenaltyEvent,
  TimerTickEvent,
  WordNewEvent,
} from '@atoyr/shared';
import { INestApplication } from '@nestjs/common';
import request from 'supertest';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { collectSSEEvents, createSSEClient, createTestApp, startGameSession, waitForSSEEvent } from '../setup';

describe('SSE Game Flow (e2e)', () => {
  let app: INestApplication;
  let baseUrl: string;

  beforeAll(async () => {
    app = await createTestApp();
    await app.listen(0);
    baseUrl = `http://127.0.0.1:${app.getHttpServer().address().port}`;
  });

  afterAll(async () => {
    await app.close();
  });

  describe('SSE Connection', () => {
    it('should emit session:started event on connection', async () => {
      const { sessionId } = await startGameSession(app);

      const es = createSSEClient(`${baseUrl}/api/game/sse/${sessionId}`);

      try {
        const data = await waitForSSEEvent<SessionStartedEvent>(es, 'session:started');
        expect(data.type).toBe('session:started');
        expect(data.sessionId).toBe(sessionId);
        expect(data.autoVoice).toBe(false);
      } finally {
        es.close();
      }
    });

    it('should emit session:started with extended time when autoVoice is true', async () => {
      const response = await request(app.getHttpServer()).post('/api/game/start').send({ autoVoice: true }).expect(201);

      const { sessionId } = { sessionId: response.body.sessionId };

      const es = createSSEClient(`${baseUrl}/api/game/sse/${sessionId}`);

      try {
        const data = await waitForSSEEvent<SessionStartedEvent>(es, 'session:started');
        expect(data.remainingSeconds).toBe(GAME_DURATION_SECONDS + AUTO_VOICE_EXTRA_SECONDS);
      } finally {
        es.close();
      }
    });

    it('should emit word:new event after session:started', async () => {
      const { sessionId } = await startGameSession(app);

      const es = createSSEClient(`${baseUrl}/api/game/sse/${sessionId}`);

      try {
        await waitForSSEEvent<SessionStartedEvent>(es, 'session:started');
        const data = await waitForSSEEvent<WordNewEvent>(es, 'word:new');

        expect(data.type).toBe('word:new');
        expect(data.scrambled).toBeDefined();
        expect(data.scrambled.length).toBe(5);
        expect(data.definition).toBeDefined();
        expect(data.token).toBeDefined();
        expect(data.wordIndex).toBe(1);
      } finally {
        es.close();
      }
    });

    it('should emit timer:tick events every second', async () => {
      const { sessionId } = await startGameSession(app);

      const es = createSSEClient(`${baseUrl}/api/game/sse/${sessionId}`);

      try {
        const events = await collectSSEEvents(es, ['timer:tick'], 3500);

        expect(events['timer:tick'].length).toBeGreaterThanOrEqual(2);

        const firstTick = events['timer:tick'][0] as TimerTickEvent;
        const secondTick = events['timer:tick'][1] as TimerTickEvent;

        expect(firstTick.remainingSeconds).toBe(GAME_DURATION_SECONDS - 1);
        expect(secondTick.remainingSeconds).toBe(GAME_DURATION_SECONDS - 2);
      } finally {
        es.close();
      }
    });

    it('should return 404 for non-existent session', async () => {
      const fakeSessionId = '00000000-0000-0000-0000-000000000000';
      const es = createSSEClient(`${baseUrl}/api/game/sse/${fakeSessionId}`);

      try {
        await expect(waitForSSEEvent(es, 'session:started', 2000)).rejects.toThrow();
      } finally {
        es.close();
      }
    });
  });

  describe('Answer Submission', () => {
    it('should emit timer:penalty event on wrong answer', async () => {
      const { sessionId } = await startGameSession(app);

      const es = createSSEClient(`${baseUrl}/api/game/sse/${sessionId}`);

      try {
        await waitForSSEEvent<SessionStartedEvent>(es, 'session:started');
        const wordData = await waitForSSEEvent<WordNewEvent>(es, 'word:new');

        // Submit wrong answer
        const penaltyPromise = waitForSSEEvent<TimerPenaltyEvent>(es, 'timer:penalty');

        await request(app.getHttpServer())
          .post('/api/game/answer')
          .send({
            sessionId,
            token: wordData.token,
            answer: 'XXXXX',
          })
          .expect(200);

        const penaltyData = await penaltyPromise;

        expect(penaltyData.type).toBe('timer:penalty');
        expect(penaltyData.penaltySeconds).toBe(1);
      } finally {
        es.close();
      }
    });
  });
});
