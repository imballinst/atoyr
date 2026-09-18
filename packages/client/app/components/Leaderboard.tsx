import { formatDate } from 'date-fns';
import { Loader2Icon } from 'lucide-react';
import { useId, useState, type JSX } from 'react';

import type { LeaderboardPeriod, SessionMode, SessionTopicLeaderboard } from '~/api/gen';
import { getFinalScore } from '~/lib/game';

import { useLeaderboard, useLeaderboardPercentile, type LeaderboardSettings } from '../api/hooks';

interface PeriodButtonGroupProps {
  period: LeaderboardPeriod;
  onChange: (period: LeaderboardPeriod) => void;
}

const LEADERBOARD_OPTIONS: Array<{ label: string; value: SessionTopicLeaderboard }> = [
  { value: 'english-words', label: 'English Words' },
  { value: 'indonesian-politician-quotes', label: 'Indonesian Politician Quotes' },
  { value: 'english-words-july-2026', label: 'English Words, July 2026' },
];

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

  const pretext = buildPretext({ percentile, rank });
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

      <div className="flex gap-3 items-start flex-wrap">
        {!settings?.mode && (
          <div className="flex flex-col gap-1 text-dark-text-secondary">
            <label htmlFor={modeId} className="text-xs">
              Mode
            </label>

            <select
              id={modeId}
              className={
                'text-xs border border-dark-border-primary rounded px-2 py-1 bg-transparent' +
                (isIndonesianTopic ? ' cursor-not-allowed text-dark-text-muted' : '')
              }
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
          <div className="flex flex-col gap-1 text-dark-text-secondary">
            <label htmlFor={topicId} className="text-xs">
              Topic
            </label>

            <select
              id={topicId}
              className="text-xs border border-dark-border-primary rounded px-2 py-1 bg-transparent"
              onChange={(e) => {
                const newTopic = e.target.value as SessionTopicLeaderboard;

                setTopic(newTopic);
                if (newTopic === 'indonesian-politician-quotes') {
                  setMode('vanilla' as SessionMode);
                }
              }}
              value={topic}
            >
              {LEADERBOARD_OPTIONS.map((opt) => (
                <option value={opt.value}>{opt.label}</option>
              ))}
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
                <div key={result.id} className="flex gap-2 p-2 md:p-3 text-xs tabular-nums">
                  <div className="font-semibold text-dark-text-primary min-w-6">#{i + 1}</div>
                  <div className="font-semibold text-dark-text-primary font-mono flex gap-2 items-end">
                    {result.id}

                    <span className="text-[10px] text-gray-400">
                      {result.isSessionSameAsCurrentUser ? '(you, just now)' : formatDate(result.timestamp, 'yyyy/MM/dd HH:mm')}
                    </span>
                  </div>
                  <div className="flex flex-1 gap-x-3 font-mono">
                    <div className="flex-1 text-right font-semibold text-dark-text-primary">
                      {getFinalScore(result.score, result.totalAttempts)}
                    </div>
                    <div className="font-semibold text-dark-interactive-success text-right min-w-7">{Math.round(result.accuracy)}%</div>
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

function buildPretext({ percentile, rank }: { percentile?: number; rank?: number }) {
  if (percentile === undefined || rank === undefined) return '';

  return (
    <span>
      Your result was better than <Percentile value={percentile} />% of players!
    </span>
  );
}

function Percentile({ value }: { value: number }) {
  const formatted = value === Math.trunc(value) ? value : value.toFixed(2);
  return <span className={`font-bold tabular-nums ${parseColor(value)}`}>{formatted}</span>;
}

function parseColor(percentile: number) {
  if (percentile >= 99) return 'text-amber-400';
  if (percentile >= 95) return 'text-fuchsia-400';
  if (percentile >= 75) return 'text-purple-400';
  if (percentile >= 50) return 'text-blue-400';
  if (percentile >= 25) return 'text-emerald-400';
  return 'text-slate-400';
}
