import { Loader2Icon } from 'lucide-react';
import type { JSX } from 'react';

import { getFinalScore } from '~/lib/game';

import { useLeaderboard } from '../api/hooks';

export function Leaderboard({ limit, HeadingComponent }: { limit?: number; HeadingComponent: keyof JSX.IntrinsicElements }) {
  const leaderboardQuery = useLeaderboard(undefined, limit);
  const leaderboardEntries = leaderboardQuery.data?.entries;

  return (
    <div className="flex flex-col h-full w-full">
      <HeadingComponent className="text-lg font-semibold mb-4 text-dark-text-primary">Leaderboard</HeadingComponent>

      <div>
        {leaderboardQuery.error ? (
          <div>Error loading leaderboard</div>
        ) : leaderboardQuery.isFetching ? (
          <Loader2Icon className="animate-spin text-dark-text-primary" />
        ) : leaderboardEntries ? (
          <div className="bg-dark-bg-tertiary rounded">
            {leaderboardEntries.map((result, i) => (
              <div key={result.id} className="flex gap-2 p-3 text-xs tabular-nums">
                <div className="font-semibold text-dark-text-primary min-w-6">#{i + 1}</div>
                <div className="font-semibold text-dark-text-primary font-mono">
                  {result.id} {result.isSessionSameAsCurrentUser ? '(You)' : ''}
                </div>
                <div className="flex flex-1 gap-x-3 font-mono">
                  <div className="flex-1 text-right font-semibold text-dark-text-primary">
                    {getFinalScore(result.score, result.totalAttempts)} attempts
                  </div>
                  <div className="font-semibold text-dark-interactive-success text-right min-w-11">{result.accuracy}%</div>
                </div>
              </div>
            ))}
          </div>
        ) : null}
      </div>
    </div>
  );
}
