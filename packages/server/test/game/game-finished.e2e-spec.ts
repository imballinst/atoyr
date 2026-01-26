import { GameFinishedEvent } from '@atoyr/shared';
import { INestApplication } from '@nestjs/common';
import request from 'supertest';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { collectSSEEvents, createSSEClient, createTestApp, startGameSessionWithCustomTime } from '../setup';

describe('Game Finished (e2e)', () => {
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

  describe('Game Completion', () => {
    it('should emit game:finished event when timer reaches 0', async () => {
      // Create a session with minimal time for faster test
      const { sessionId } = await startGameSessionWithCustomTime(app, 2);

      const es = createSSEClient(`${baseUrl}/api/game/sse/${sessionId}`);

      try {
        const events = await collectSSEEvents(es, ['session:started', 'word:new', 'timer:tick', 'game:finished'], 5000);

        expect(events['game:finished'].length).toBe(1);

        const finishedData = events['game:finished'][0] as GameFinishedEvent;

        expect(finishedData.type).toBe('game:finished');
        expect(finishedData.score).toBeDefined();
        expect(finishedData.totalAttempts).toBeDefined();
        expect(finishedData.accuracy).toBeDefined();
        expect(finishedData.resultId).toBeDefined();
      } finally {
        es.close();
      }
    }, 10000);

    it('should save result to database on game finish', async () => {
      const { sessionId } = await startGameSessionWithCustomTime(app, 2);

      const es = createSSEClient(`${baseUrl}/api/game/sse/${sessionId}`);

      try {
        const events = await collectSSEEvents(es, ['game:finished'], 5000);

        const finishedData = events['game:finished'][0] as GameFinishedEvent;

        // Verify result is in leaderboard
        const leaderboard = await request(app.getHttpServer()).get('/api/leaderboard').expect(200);

        const savedResult = leaderboard.body.entries.find((e: any) => e.id === finishedData.resultId);

        expect(savedResult).toBeDefined();
        expect(savedResult.score).toBe(finishedData.score);
      } finally {
        es.close();
      }
    }, 10000);

    it('should reject answers after game is finished', async () => {
      const { sessionId } = await startGameSessionWithCustomTime(app, 2);

      const es = createSSEClient(`${baseUrl}/api/game/sse/${sessionId}`);

      try {
        const events = await collectSSEEvents(es, ['word:new', 'game:finished'], 5000);

        const wordData = events['word:new'][0];

        // Try to submit answer after game finished
        await request(app.getHttpServer())
          .post('/api/game/answer')
          .send({
            sessionId,
            token: (wordData as any).token || 'invalid',
            answer: 'BOARD',
          })
          .expect(400);
      } finally {
        es.close();
      }
    }, 10000);
  });
});
