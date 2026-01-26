import { INestApplication } from '@nestjs/common';
import request from 'supertest';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { createTestApp } from '../setup';

describe('Game Start (e2e)', () => {
  let app: INestApplication;

  beforeAll(async () => {
    app = await createTestApp();
  });

  afterAll(async () => {
    await app.close();
  });

  describe('POST /api/game/start', () => {
    it('should create a new game session', async () => {
      const response = await request(app.getHttpServer()).post('/api/game/start').send({ autoVoice: false }).expect(201);

      expect(response.body).toHaveProperty('sessionId');
      expect(response.body).toHaveProperty('expiresAt');
      expect(typeof response.body.sessionId).toBe('string');
      expect(response.body.sessionId).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i);
    });

    it('should create session with autoVoice enabled', async () => {
      const response = await request(app.getHttpServer()).post('/api/game/start').send({ autoVoice: true }).expect(201);

      expect(response.body).toHaveProperty('sessionId');
    });

    it('should default autoVoice to false if not provided', async () => {
      const response = await request(app.getHttpServer()).post('/api/game/start').send({}).expect(201);

      expect(response.body).toHaveProperty('sessionId');
    });

    it('should return expiry time in the future', async () => {
      const response = await request(app.getHttpServer()).post('/api/game/start').send({}).expect(201);

      const now = Date.now();
      expect(response.body.expiresAt).toBeGreaterThan(now);
    });
  });
});
