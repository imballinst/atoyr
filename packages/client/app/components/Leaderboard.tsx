import { getFinalScore } from '~/lib/game';
import { useLeaderboard } from '../api/hooks';
import { Loader2Icon } from 'lucide-react'
import type { JSX } from 'react';

export function Leaderboard({ limit, HeadingComponent }: { limit?: number, HeadingComponent: keyof JSX.IntrinsicElements }) {
  const leaderboardQuery = useLeaderboard(undefined, limit);
  const leaderboardEntries = leaderboardQuery.data?.entries;

  return (
    <div className="flex flex-col h-full w-full gap-2">
      <HeadingComponent className="text-lg font-semibold mb-4 text-dark-text-primary">Leaderboard</HeadingComponent>

      <div>
        {leaderboardQuery.error ? (
          <div>Error loading leaderboard</div>
        ) : leaderboardQuery.isFetching ? (
          <Loader2Icon className='animate-spin text-dark-text-primary' />
        ) : leaderboardEntries ? (
          leaderboardEntries.map((result, i) => (
            <div key={result.id} className="flex gap-2 p-3 bg-dark-bg-tertiary rounded text-xs">
              <div className="font-semibold text-dark-text-primary">#{i + 1}</div>
              <div className="font-semibold text-dark-text-primary font-mono">
                {result.id} {result.isSessionSameAsCurrentUser ? '(You)' : ''}
              </div>
              <div className="flex flex-1 gap-x-3 font-mono">
                <div className="flex-1 text-right font-semibold text-dark-text-primary">
                  {getFinalScore(result.score, result.totalAttempts)} attempts
                </div>
                <div className="font-semibold text-dark-interactive-success text-right min-w-7">{result.accuracy}%</div>
              </div>
            </div>
          ))
        ) : null}
      </div>
    </div>
  );
}
