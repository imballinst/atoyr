import { Loader2Icon } from 'lucide-react';
import { useId, useState, type JSX } from 'react';

import type { SessionMode, SessionTopic } from '~/api/gen';
import { getFinalScore } from '~/lib/game';

import { useLeaderboard, type LeaderboardSettings } from '../api/hooks';

export function Leaderboard({
  settings,
  limit,
  HeadingComponent,
}: {
  settings?: LeaderboardSettings;
  limit?: number;
  HeadingComponent: keyof JSX.IntrinsicElements;
}) {
  const [mode, setMode] = useState(settings?.mode);
  const modeId = useId();
  const [topic, setTopic] = useState(settings?.topic);
  const topicId = useId();

  const leaderboardQuery = useLeaderboard({ mode, topic }, undefined, limit);
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

  const isIndonesianTopic = topic === 'indonesian-politician-quotes';

  return (
    <div className="flex flex-col h-full w-full gap-y-2">
      <HeadingComponent className="text-lg font-semibold text-dark-text-primary">Leaderboard</HeadingComponent>

      <p className="text-xs text-dark-text-secondary text-left">
        {pretext} Order priority: more correct answers → more accuracy → earlier record time.
      </p>

      <div className="flex gap-2">
        {!settings?.mode && (
          <div className="text-dark-text-secondary">
            <label htmlFor={modeId} className="sr-only">
              Mode
            </label>

            <select
              className={'text-sm text-right' + (isIndonesianTopic ? ' cursor-not-allowed text-dark-text-muted' : '')}
              onChange={(e) => {
                setMode(e.target.value as SessionMode);
              }}
              disabled={isIndonesianTopic}
              value={mode}
            >
              <option value="vanilla">Vanilla</option>
              <option value="blind">Blind</option>
            </select>
          </div>
        )}

        {!settings?.topic && (
          <div className="text-dark-text-secondary">
            <label htmlFor={topicId} className="sr-only">
              Topic
            </label>

            <select
              className="text-sm text-right"
              onChange={(e) => {
                const newTopic = e.target.value as SessionTopic;

                setTopic(newTopic);
                if (newTopic === 'indonesian-politician-quotes') {
                  setMode('vanilla' as SessionMode);
                }
              }}
              value={topic}
            >
              <option value="english-words">English Words</option>
              <option value="indonesian-politician-quotes">Indonesian Quotes</option>
            </select>
          </div>
        )}
      </div>

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
