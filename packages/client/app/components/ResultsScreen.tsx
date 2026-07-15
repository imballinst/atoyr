import type { ReactNode } from 'react';

import { useLeaderboardPercentile } from '~/api/hooks';

import { type GameState } from '../lib/game';
import { Leaderboard } from './Leaderboard';
import { SettingsModal } from './SettingsModal';

interface ResultsScreenProps extends Pick<
  GameState,
  'score' | 'totalAttempts' | 'correctAttemptTimestamps' | 'currentWord' | 'lastWordAnswer'
> {
  onPlayAgain: () => void;
  onBackToHome: () => void;
}

export function ResultsScreen({
  score,
  totalAttempts,
  correctAttemptTimestamps,
  onPlayAgain,
  onBackToHome,
  currentWord,
  lastWordAnswer,
}: ResultsScreenProps) {
  const { data } = useLeaderboardPercentile();

  const accuracy = totalAttempts > 0 ? Math.trunc((score / totalAttempts) * 10000) / 100 : 0;
  const longestStreak = Math.max(...correctAttemptTimestamps.map((attempts) => attempts.length), 0);

  return (
    <>
      <h1 className="text-4xl font-bold mb-2 text-dark-text-primary">Game Over!</h1>

      <div className="text-dark-text-primary border p-2 rounded border-dark-bg-tertiary text-sm mb-4">
        Last word: {currentWord?.scrambled} → <span className="font-bold">{lastWordAnswer}</span>
      </div>

      <div className="flex flex-col gap-2 mb-6 w-full text-center">
        <div className="border-dark-bg-tertiary text-dark-text-primary p-2 sm:p-4 rounded-lg col-span-5 text-sm">
          Your result was better than <Percentile value={data?.percentile} />% other players!
        </div>
        <div className="grid grid-cols-6 gap-2">
          <div className="bg-dark-bg-tertiary p-2 sm:p-4 rounded-lg col-span-3 md:col-span-2">
            <Stat label="Score">
              {score} / {totalAttempts}
            </Stat>
          </div>
          <div className="bg-dark-bg-tertiary p-2 sm:p-4 rounded-lg col-span-3 md:col-span-2">
            <Stat label="Accuracy">{accuracy}%</Stat>
          </div>
          <div className="bg-dark-bg-tertiary p-2 sm:p-4 rounded-lg col-span-6 md:col-span-2">
            <Stat label="Best streak">{longestStreak}</Stat>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-6 gap-2 w-full">
        <button
          type="button"
          onClick={onBackToHome}
          className="col-span-2 py-2 px-4 text-sm font-medium text-dark-text-secondary border border-dark-border-primary rounded transition duration-200 hover:bg-dark-bg-tertiary hover:text-dark-text-primary"
          data-ga-label="ga-back-to-home-button"
        >
          Back to home
        </button>
        <div className="col-span-4 flex gap-2">
          <div className="flex-1">
            <SettingsModal />
          </div>
          <button
            type="button"
            onClick={onPlayAgain}
            className="flex-1 py-3 px-4 text-sm font-semibold bg-dark-interactive-primary text-white rounded transition duration-200 hover:-translate-y-0.5 hover:shadow-lg active:translate-y-0 shadow hover:bg-dark-interactive-hover"
            data-ga-label="ga-play-again-button"
          >
            Play Again
          </button>
        </div>
      </div>

      <div className="pt-6 w-full min-h-[270px]">
        <Leaderboard limit={5} HeadingComponent="h2" />
      </div>
    </>
  );
}

function Percentile({ value }: { value: number | undefined }) {
  if (value === undefined) {
    return <span className="animate-pulse">--</span>;
  }

  const formatted = value === Math.trunc(value) ? value : value.toFixed(2);
  return <span className={`font-bold tabular-nums ${parseColor(value)}`}>{formatted}</span>;
}

function Stat({ label, children }: { label: ReactNode; children: ReactNode }) {
  return (
    <>
      <div className="text-xs text-dark-text-tertiary mb-2 font-medium">{label}</div>
      <div className="text-3xl font-bold text-dark-interactive-success">{children}</div>
    </>
  );
}

function parseColor(percentile: number) {
  if (percentile >= 99) return 'text-amber-400';
  if (percentile >= 95) return 'text-fuchsia-400';
  if (percentile >= 75) return 'text-purple-400';
  if (percentile >= 50) return 'text-blue-400';
  if (percentile >= 25) return 'text-emerald-400';
  return 'text-slate-400';
}
