import { Loader2Icon } from 'lucide-react';
import { useId, useState, type JSX } from 'react';

import type { LeaderboardPeriod, SessionMode, SessionTopic } from '~/api/gen';
import { getFinalScore } from '~/lib/game';

import { useLeaderboard, useLeaderboardPercentile, type LeaderboardSettings } from '../api/hooks';

interface PeriodButtonGroupProps {
  period: LeaderboardPeriod;
  onChange: (period: LeaderboardPeriod) => void;
}

const PERIOD_OPTIONS: { value: LeaderboardPeriod; label: string }[] = [
  { value: 'alltime', label: 'All-time' },
  { value: 'monthly', label: 'This month' },
];

function PeriodButtonGroup({ period, onChange }: PeriodButtonGroupProps) {
  return (
    <div
      role="radiogroup"
      aria-label="Leaderboard period"
      className="inline-flex border border-dark-border-primary rounded overflow-hidden text-xs"
    >
      {PERIOD_OPTIONS.map((option) => {
        const isActive = period === option.value;
        return (
          <button
            key={option.value}
            type="button"
            role="radio"
            aria-checked={isActive}
            onClick={() => onChange(option.value)}
            className={
              'px-3 py-1 transition duration-200 ' +
              (isActive
                ? 'bg-dark-interactive-primary text-white'
                : 'bg-transparent text-dark-text-secondary hover:bg-dark-bg-tertiary hover:text-dark-text-primary')
            }
          >
            {option.label}
          </button>
        );
      })}
    </div>
  );
}

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
  const [period, setPeriod] = useState<LeaderboardPeriod>('alltime');

  const leaderboardQuery = useLeaderboard({ mode, topic }, period, undefined, limit);
  const percentileQuery = useLeaderboardPercentile({ mode, topic }, period);
  const leaderboardEntries = leaderboardQuery.data?.entries;
  const percentile = percentileQuery.data?.percentile;
  const rank = percentileQuery.data?.rank;

  const userEntryIndex = leaderboardEntries?.findIndex((entry) => entry.isSessionSameAsCurrentUser) ?? -1;
  const userInPreview = userEntryIndex > -1;

  const pretext = buildPretext({ percentile, rank, userInPreview });
  const isIndonesianTopic = topic === 'indonesian-politician-quotes';

  return (
    <div className="flex flex-col h-full w-full gap-y-2">
      <div className="flex items-center justify-between gap-2 flex-wrap">
        <HeadingComponent className="text-lg font-semibold text-dark-text-primary">Leaderboard</HeadingComponent>
        <PeriodButtonGroup period={period} onChange={setPeriod} />
      </div>

      <p className="text-xs text-dark-text-secondary text-left">
        {pretext} Order priority: more correct answers → more accuracy → earlier record time.
      </p>

      <div className="flex gap-3 items-center flex-wrap">
        {!settings?.mode && (
          <div className="flex items-center gap-1 text-dark-text-secondary">
            <label htmlFor={modeId} className="text-xs">
              Mode
            </label>

            <select
              id={modeId}
              className={'text-xs' + (isIndonesianTopic ? ' cursor-not-allowed text-dark-text-muted' : '')}
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
          <div className="flex items-center gap-1 text-dark-text-secondary">
            <label htmlFor={topicId} className="text-xs">
              Topic
            </label>

            <select
              id={topicId}
              className="text-xs"
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

function buildPretext({ percentile, rank, userInPreview }: { percentile?: number; rank?: number; userInPreview: boolean }): string {
  if (percentile === undefined || rank === undefined) return '';

  const formatted = percentile === Math.trunc(percentile) ? percentile.toString() : percentile.toFixed(2);
  if (userInPreview) {
    return `Your result was better than ${formatted}% of players! You also got a placement in leaderboard #${rank}.`;
  }
  return `Your result was better than ${formatted}% of players. Unfortunately, you didn't make it to the leaderboard.`;
}
