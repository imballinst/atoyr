import { Loader2Icon } from 'lucide-react';
import { useId, useState, type JSX } from 'react';

import type { SessionMode } from '~/api/gen';
import { getFinalScore } from '~/lib/game';

import { useLeaderboard } from '../api/hooks';

export function Leaderboard({
  mode: modeProps,
  limit,
  HeadingComponent,
}: {
  mode?: SessionMode;
  limit?: number;
  HeadingComponent: keyof JSX.IntrinsicElements;
}) {
  const [mode, setMode] = useState(modeProps);
  const modeId = useId();

  const leaderboardQuery = useLeaderboard(mode, undefined, limit);
  const leaderboardEntries = leaderboardQuery.data?.entries;
  let pretext = '';

  if (HeadingComponent === 'h2' && leaderboardEntries) {
    // h2 means inside the game result screen.
    const idx = leaderboardEntries.findIndex((item) => item.isSessionSameAsCurrentUser);
    if (idx > -1) {
      pretext = `Your result ranked ${idx + 1}!`;
    } else {
      pretext = "Unfortunately, your result didn't make it.";
    }
  }

  return (
    <div className="flex flex-col h-full w-full gap-y-2">
      <HeadingComponent className="text-lg font-semibold text-dark-text-primary">Leaderboard</HeadingComponent>

      <p className="text-xs text-dark-text-secondary text-left">
        {pretext} Order priority: more correct answers → more accuracy → earlier record time.
      </p>

      {!modeProps && (
        <div className="text-dark-text-secondary">
          <label htmlFor={modeId} className="sr-only">
            Mode
          </label>

          <select
            className="text-sm"
            onChange={(e) => {
              setMode(e.target.value as SessionMode);
            }}
            value={mode}
          >
            <option value="vanilla">Vanilla</option>
            <option value="blind">Blind</option>
          </select>
        </div>
      )}

      <div className="text-dark-text-secondary text-sm">
        {leaderboardQuery.error ? (
          <div>Error loading leaderboard</div>
        ) : leaderboardQuery.isFetching ? (
          <Loader2Icon className="animate-spin text-dark-text-primary" />
        ) : leaderboardEntries ? (
          leaderboardEntries.length === 0 ? (
            <div>No leaderboard entries yet.</div>
          ) : (
            <div className="bg-dark-bg-tertiary rounded">
              {leaderboardEntries.map((result, i) => (
                <div key={result.id} className="flex gap-2 p-3 text-xs tabular-nums">
                  <div className="font-semibold text-dark-text-primary min-w-6">#{i + 1}</div>
                  <div className="font-semibold text-dark-text-primary font-mono">
                    {result.id} {result.isSessionSameAsCurrentUser ? '(you, last game)' : ''}
                  </div>
                  <div className="flex flex-1 gap-x-3 font-mono">
                    <div className="flex-1 text-right font-semibold text-dark-text-primary">
                      {getFinalScore(result.score, result.totalAttempts)}
                    </div>
                    <div className="font-semibold text-dark-interactive-success text-right min-w-11">{result.accuracy}%</div>
                  </div>
                </div>
              ))}
            </div>
          )
        ) : null}
      </div>
    </div>
  );
}
