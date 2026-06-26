import { getFinalScore, type GameState } from '../lib/game';
import { Leaderboard } from './Leaderboard';

interface ResultsScreenProps extends Pick<GameState, 'score' | 'totalAttempts' | 'correctAttemptTimestamps'> {
  onPlayAgain: () => void;
}

export function ResultsScreen({ score, totalAttempts, correctAttemptTimestamps, onPlayAgain }: ResultsScreenProps) {
  const accuracy = totalAttempts > 0 ? (score / totalAttempts) * 100 : 0;
  const longestStreak = Math.max(...correctAttemptTimestamps.map((attempts) => attempts.length), 0);

  return (
    <>
      <h1 className="text-4xl font-bold mb-6 text-dark-text-primary">Game Over!</h1>

      <div className="flex flex-col gap-2 mb-3 w-full text-center">
        <div className="grid grid-cols-5 gap-2">
          <div className="bg-dark-bg-tertiary p-2 sm:p-4 rounded-lg col-span-3">
            <div className="text-xs text-dark-text-tertiary mb-2 font-medium">Score</div>
            <div className="text-3xl font-bold text-dark-text-primary">{getFinalScore(score, totalAttempts)}</div>
          </div>
          <div className="bg-dark-bg-tertiary p-2 sm:p-4 rounded-lg col-span-2">
            <div className="text-xs text-dark-text-tertiary mb-2 font-medium">Accuracy</div>
            <div className="text-3xl font-bold text-dark-interactive-success">{accuracy}%</div>
          </div>
        </div>
        <div className="grid grid-cols-8 gap-2">
          <div className="bg-dark-bg-tertiary p-2 sm:p-4 rounded-lg col-span-8">
            <div className="text-xs text-dark-text-tertiary mb-2 font-medium">Best streak</div>
            <div className="text-3xl font-bold text-dark-text-primary">{longestStreak}</div>
          </div>
        </div>
      </div>

      <button
        onClick={onPlayAgain}
        className="w-full py-3 px-6 text-base font-semibold bg-dark-interactive-primary text-white rounded-lg transition duration-200 hover:-translate-y-0.5 hover:shadow-lg active:translate-y-0 shadow mb-6 hover:bg-dark-interactive-hover"
      >
        Play Again
      </button>

      <div className="mt-6 pt-6 border-t border-dark-border-secondary w-full">
        <Leaderboard limit={5} HeadingComponent="h2" />
      </div>
    </>
  );
}
