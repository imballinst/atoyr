import { z } from 'zod';
import { WORD_LENGTH } from './types';

export const startGameRequestSchema = z.object({
  autoVoice: z.boolean().optional().default(false),
});

export const submitAnswerRequestSchema = z.object({
  sessionId: z.string().uuid(),
  token: z.string().min(1),
  answer: z.string().length(WORD_LENGTH),
});

export const uuidSchema = z.string().uuid();

export const leaderboardQuerySchema = z.object({
  limit: z.coerce.number().int().min(1).max(100).optional().default(10),
  offset: z.coerce.number().int().min(0).optional().default(0),
});

export type StartGameRequestDto = z.infer<typeof startGameRequestSchema>;
export type SubmitAnswerRequestDto = z.infer<typeof submitAnswerRequestSchema>;
export type LeaderboardQueryDto = z.infer<typeof leaderboardQuerySchema>;
