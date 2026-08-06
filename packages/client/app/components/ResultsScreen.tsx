import { Fragment, type ReactNode } from 'react';

import { useLeaderboardPercentile } from '~/api/hooks';
import type { LatestSchema } from '~/lib/settings';

import { type GameState } from '../lib/game';
import { Leaderboard } from './Leaderboard';
import { SettingsModal } from './SettingsModal';
import { StatsBar } from './StatsBar';

interface ResultsScreenProps extends Pick<
  GameState,
  'score' | 'totalAttempts' | 'correctAttemptTimestamps' | 'currentWord' | 'lastWordAnswer' | 'lastWordDefinition' | 'settings'
> {
  onPlayAgain: () => void;
  onBackToHome: () => void;
  onUpdateSettings: (config: Partial<LatestSchema>) => void;
}

export function ResultsScreen({
  score,
  totalAttempts,
  correctAttemptTimestamps,
  onPlayAgain,
  onBackToHome,
  currentWord,
  lastWordAnswer,
  lastWordDefinition,
  settings,
  onUpdateSettings,
}: ResultsScreenProps) {
  const { data } = useLeaderboardPercentile(settings);

  const accuracy = totalAttempts > 0 ? Math.trunc((score / totalAttempts) * 10000) / 100 : 0;
  const longestStreak = Math.max(...correctAttemptTimestamps.map((attempts) => attempts.length), 0);

  const isIndonesianQuotes = settings.topic === 'indonesian-politician-quotes';

  return (
    <>
      <div className="text-dark-text-primary border p-2 rounded border-dark-bg-tertiary text-sm mb-4">
        {lastWordAnswer &&
          (isIndonesianQuotes && lastWordDefinition ? (
            <p>
              Last quote: <span className="italic text-dark-text-tertiary">{highlightAnswer(lastWordDefinition, lastWordAnswer)}</span>
            </p>
          ) : (
            <p>
              Last word: {currentWord?.scrambled} → <span className="font-bold">{lastWordAnswer}</span>
            </p>
          ))}
      </div>

      <div className="flex flex-col gap-2 mb-6 w-full text-center">
        <div className="border-dark-bg-tertiary text-dark-text-primary rounded-lg col-span-5 text-sm mb-2">
          Game over: your result was better than <Percentile value={data?.percentile} />% other players!
        </div>
        <StatsBar
          stats={[
            { label: 'Score', value: `${score} / ${totalAttempts}` },
            { label: 'Accuracy', value: `${accuracy}%` },
            { label: 'Best streak', value: longestStreak },
          ]}
        />
      </div>

      <div className="grid grid-cols-10 gap-2 w-full">
        <button
          type="button"
          onClick={onBackToHome}
          className="col-span-4 py-2 px-4 text-sm font-medium text-dark-text-secondary border border-dark-border-primary rounded transition duration-200 hover:bg-dark-bg-tertiary hover:text-dark-text-primary"
          data-ga-label="ga-back-to-home-button"
        >
          Back to home
        </button>
        <div className="col-span-6 flex">
          <SettingsModal
            triggerText={'⚙️'}
            triggerClassnames="border-r-0 rounded-r-none"
            settings={settings}
            updateSettings={onUpdateSettings}
          />

          <button
            type="button"
            onClick={onPlayAgain}
            className="flex-1 py-3 px-4 text-sm font-semibold bg-dark-interactive-primary text-white rounded rounded-l-none hover:shadow-lg active:translate-y-0 shadow hover:bg-dark-interactive-hover"
            data-ga-label="ga-play-again-button"
          >
            Play Again
          </button>
        </div>
      </div>

      <div className="pt-6 w-full min-h-[270px]">
        <Leaderboard settings={settings} limit={5} HeadingComponent="h2" />
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

function parseColor(percentile: number) {
  if (percentile >= 99) return 'text-amber-400';
  if (percentile >= 95) return 'text-fuchsia-400';
  if (percentile >= 75) return 'text-purple-400';
  if (percentile >= 50) return 'text-blue-400';
  if (percentile >= 25) return 'text-emerald-400';
  return 'text-slate-400';
}

function highlightAnswer(definition: string, answer: string): ReactNode {
  const parts = definition.split(answer);
  return parts.map((part, i) => (
    <Fragment key={i}>
      {part}
      {i < parts.length - 1 && <strong className="underline">{answer}</strong>}
    </Fragment>
  ));
}
