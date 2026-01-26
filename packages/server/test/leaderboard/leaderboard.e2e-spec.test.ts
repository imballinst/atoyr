import { INestApplication } from '@nestjs/common';
import request from 'supertest';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { createTestApp } from '../setup';

describe('Leaderboard (e2e)', () => {
  let app: INestApplication;

  beforeAll(async () => {
    app = await createTestApp();
  });

  afterAll(async () => {
    await app.close();
  });

  describe('GET /api/leaderboard', () => {
    it('should return leaderboard structure', async () => {
      const response = await request(app.getHttpServer()).get('/api/leaderboard').expect(200);

      expect(response.body).toHaveProperty('entries');
      expect(response.body).toHaveProperty('totalGames');
      expect(response.body).toHaveProperty('lastUpdated');
      expect(Array.isArray(response.body.entries)).toBe(true);
    });

    it('should return leaderboard sorted by score descending', async () => {
      const response = await request(app.getHttpServer()).get('/api/leaderboard').expect(200);

      const entries = response.body.entries;
      for (let i = 1; i < entries.length; i++) {
        expect(entries[i - 1].score).toBeGreaterThanOrEqual(entries[i].score);
      }
    });

    it('should respect limit parameter', async () => {
      const response = await request(app.getHttpServer()).get('/api/leaderboard?limit=5').expect(200);

      expect(response.body.entries.length).toBeLessThanOrEqual(5);
    });

    it('should respect offset parameter', async () => {
      const fullResponse = await request(app.getHttpServer()).get('/api/leaderboard?limit=10').expect(200);

      const offsetResponse = await request(app.getHttpServer()).get('/api/leaderboard?limit=5&offset=5').expect(200);

      if (fullResponse.body.entries.length > 5) {
        expect(offsetResponse.body.entries[0]).toEqual(fullResponse.body.entries[5]);
      }
    });

    it('should include rank in entries', async () => {
      const response = await request(app.getHttpServer()).get('/api/leaderboard').expect(200);

      response.body.entries.forEach((entry: any, index: number) => {
        expect(entry.rank).toBe(index + 1);
      });
    });

    it('should cap limit at 100', async () => {
      const response = await request(app.getHttpServer()).get('/api/leaderboard?limit=200').expect(200);

      expect(response.body.entries.length).toBeLessThanOrEqual(100);
    });

    it('should default limit to 10', async () => {
      const response = await request(app.getHttpServer()).get('/api/leaderboard').expect(200);

      expect(response.body.entries.length).toBeLessThanOrEqual(10);
    });
  });
});
