import { GameResult, GameState } from '../types/game';

interface ResultsScreenProps extends Pick<GameState, 'score' | 'totalAttempts' | 'correctAttemptTimestamps'> {
  leaderboard: GameResult[];
  onPlayAgain: () => void;
}

export function ResultsScreen({ score, totalAttempts, correctAttemptTimestamps, leaderboard, onPlayAgain }: ResultsScreenProps) {
  const accuracy = totalAttempts > 0 ? ((score / totalAttempts) * 100).toFixed(1) : '0.0';
  const longestStreak = Math.max(...correctAttemptTimestamps.map((attempts) => attempts.length), 0);

  return (
    <div className="w-screen h-screen flex items-center justify-center p-5 bg-dark-bg-primary">
      <div className="rounded-2xl p-10 text-center w-full max-w-[430px]">
        <h1 className="text-4xl font-bold mb-6 text-dark-text-primary">Game Over!</h1>

        <div className="grid grid-cols-7 gap-3 mb-8">
          <div className="bg-dark-bg-tertiary p-4 rounded-lg col-span-2">
            <div className="text-xs text-dark-text-tertiary mb-2 font-medium">Score</div>
            <div className="text-3xl font-bold text-dark-text-primary">
              {score}/{totalAttempts}
            </div>
          </div>
          <div className="bg-dark-bg-tertiary p-4 rounded-lg col-span-3">
            <div className="text-xs text-dark-text-tertiary mb-2 font-medium">Accuracy</div>
            <div className="text-3xl font-bold text-dark-interactive-success">{accuracy}%</div>
          </div>
          <div className="bg-dark-bg-tertiary p-4 rounded-lg col-span-2">
            <div className="text-xs text-dark-text-tertiary mb-2 font-medium">Longest streak</div>
            <div className="text-3xl font-bold text-dark-text-primary">{longestStreak}</div>
          </div>
        </div>

        <button
          onClick={onPlayAgain}
          className="w-full py-3 px-6 text-base font-semibold bg-dark-interactive-primary text-white rounded-lg transition duration-200 hover:-translate-y-0.5 hover:shadow-lg active:translate-y-0 shadow mb-6 hover:bg-dark-interactive-hover"
        >
          Play Again
        </button>

        {leaderboard.length > 0 && (
          <div className="mt-6 pt-6 border-t border-dark-border-secondary">
            <h2 className="text-lg font-semibold mb-4 text-dark-text-primary">Leaderboard</h2>
            <div className="flex flex-col gap-2 max-h-52 overflow-y-auto">
              {leaderboard
                .sort((a, b) => b.score - a.score)
                .slice(0, 10)
                .map((result, i) => (
                  <div key={result.id} className="flex justify-between gap-3 p-3 bg-dark-bg-tertiary rounded text-xs">
                    <div className="font-semibold text-dark-text-primary min-w-8">#{i + 1}</div>
                    <div className="flex gap-x-1 font-mono">
                      <div className="flex-1 text-center font-semibold text-dark-text-primary">
                        {result.score}/{result.totalAttempts}
                      </div>
                      <div className="font-semibold text-dark-interactive-success w-10 text-right">
                        ({(result.accuracy * 100).toFixed(0)}%)
                      </div>
                    </div>
                  </div>
                ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
