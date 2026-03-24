// Re-export all types and constants from shared package
// This maintains backward compatibility while using shared as single source of truth
export {
  GAME_DURATION_SECONDS,
  LEADERBOARD_COOKIE_EXPIRY_DAYS,
  LEADERBOARD_COOKIE_NAME,
  WRONG_ANSWER_PENALTY_SECONDS,
  type GamePhase,
  type GameResult,
  type GameState,
  type WordEntry,
} from '@atoyr/shared';
