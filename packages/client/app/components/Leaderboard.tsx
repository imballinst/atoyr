import { getFinalScore } from '~/lib/game';

import { useLeaderboard } from '../api/hooks';

export function Leaderboard({ limit }: { limit?: number }) {
  const leaderboardQuery = useLeaderboard(undefined, limit);
  const leaderboardEntries = leaderboardQuery.data?.entries;

  return (
    <div className="flex flex-col gap-2">
      {leaderboardQuery.error ? (
        <div>Error loading leaderboard</div>
      ) : leaderboardEntries ? (
        leaderboardEntries.map((result, i) => (
          <div key={result.id} className="flex gap-2 p-3 bg-dark-bg-tertiary rounded text-xs">
            <div className="font-semibold text-dark-text-primary">#{i + 1}</div>
            <div className="font-semibold text-dark-text-primary font-mono">
              {result.id} {result.isSessionSameAsCurrentUser ? '(You)' : ''}
            </div>
            <div className="flex flex-1 gap-x-3 font-mono">
              <div className="flex-1 text-right font-semibold text-dark-text-primary">
                {getFinalScore(result.score, result.totalAttempts)}
              </div>
              <div className="font-semibold text-dark-interactive-success text-right">{result.accuracy}%</div>
            </div>
          </div>
        ))
      ) : null}
    </div>
  );
}
