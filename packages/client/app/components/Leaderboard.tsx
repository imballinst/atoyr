import { getFinalScore } from '~/lib/game';
import { useLeaderboard } from '../api/hooks';

export function Leaderboard() {
  const leaderboardQuery = useLeaderboard();

  return (
    <div className="text-dark-text-tertiary flex flex-col gap-y-4">
      <h2 className='text-center text-2xl'>Leaderboard</h2>

      {leaderboardQuery.isLoading ? (
        <div>Loading...</div>
      ) : leaderboardQuery.data ? (
        <div className="flex flex-col gap-2 max-h-52 overflow-y-auto">
          {leaderboardQuery.data.entries
            .map((result, i) => (
              <div key={result.id} className="flex gap-3 p-3 bg-dark-bg-tertiary rounded text-xs">
                <div className="font-semibold text-dark-text-primary min-w-8">#{i + 1}</div>
                <div className="font-semibold text-dark-text-primary font-mono">{result.id}</div>
                <div className="flex flex-1 gap-x-1 font-mono">
                  <div className="flex-1 text-right font-semibold text-dark-text-primary">
                    {getFinalScore(result.score, result.totalAttempts)}
                  </div>
                  <div className="font-semibold text-dark-interactive-success w-10 text-right">
                    ({(result.accuracy * 100).toFixed(0)}%)
                  </div>
                </div>
              </div>
            ))}
        </div>
      ) : leaderboardQuery.isError ? (
        <div>Error loading leaderboard</div>
      ) : null}
    </div>
  );
}
