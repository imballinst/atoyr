import { z } from 'zod';

export const Player = z.object({
  id: z.string(),
  name: z.string(),
  level: z.number().min(1),
  maxHp: z.number().min(0),
  hp: z.number().min(0),
  attack: z.number().min(0),
  defense: z.number().min(0),
  words: z.array(z.string()).optional(),
});
export type Player = z.infer<typeof Player>;

export type Role = 'attacker' | 'defender';

export const MatchState = z.object({
  id: z.string(),
  player: Player.nullable(),
  opponent: Player.nullable(),
  currentTurn: z.number().nullable(),
  currentRole: z.enum(['attacker', 'defender']).nullable(),
  currentWord: z.string().nullable(),
  turnHistory: z.array(
    z.object({
      word: z.string(),
      result: z.string(),
    }),
  ),
  rawLogs: z.array(z.string()),
});
export type MatchState = z.infer<typeof MatchState>;

const TurnResult = z.object({
  kind: z.literal('turn_result'),
  result: z.enum(['win', 'lose', 'draw']),
  damage: z.number(),
  nextTurnTimestamp: z.string(),
  nextTurnState: MatchState.pick({
    currentTurn: true,
    currentRole: true,
    opponent: true,
    player: true,
  }),
});
const TurnStart = z.object({
  kind: z.literal('turn_start'),
  turn: z.number(),
  turnEndTimestamp: z.string(),
  word: z.string(),
});

export const GameEventPayload = z.union([TurnResult, TurnStart]);
