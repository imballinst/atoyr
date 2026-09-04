import { Fragment, type ReactNode } from 'react';

import { useIsMobile } from '~/api/hooks';
import type { LatestSchema } from '~/lib/settings';

import { type GameState } from '../lib/game';
import { Leaderboard } from './Leaderboard';
import { SettingsModal } from './SettingsModal';
import { StatsBar } from './StatsBar';

interface ResultsScreenProps extends Pick<
  GameState,
  | 'score'
  | 'totalAttempts'
  | 'correctAttemptTimestamps'
  | 'currentWord'
  | 'lastWordAnswer'
  | 'lastWordDefinition'
  | 'settings'
  | 'lastWordReferences'
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
  lastWordReferences,
  settings,
  onUpdateSettings,
}: ResultsScreenProps) {
  const accuracy = totalAttempts > 0 ? Math.trunc((score / totalAttempts) * 10000) / 100 : 0;
  const longestStreak = Math.max(...correctAttemptTimestamps.map((attempts) => attempts.length), 0);
  const isMobile = useIsMobile();

  const isIndonesianQuotes = settings.topic === 'indonesian-politician-quotes';

  return (
    <>
      <div className="text-dark-text-primary border p-2 rounded border-dark-bg-tertiary text-sm mb-4 mt-2">
        {lastWordAnswer &&
          (isIndonesianQuotes && lastWordDefinition ? (
            <>
              <p>
                Game over! Last quote:{' '}
                <span className="italic text-dark-text-tertiary">{highlightAnswer(lastWordDefinition, lastWordAnswer)}</span>
              </p>

              {lastWordReferences && (
                <div className="inline-flex gap-1 mt-2 text-xs">
                  <span>Source:</span>

                  {lastWordReferences.map((ref, idx) => (
                    <span key={ref}>
                      <a href={ref} target="_blank" rel="noopener noreferrer" className="underline decoration-dotted">
                        {new URL(ref).host}
                      </a>
                      {idx + 1 < lastWordReferences.length ? ',' : null}
                    </span>
                  ))}
                </div>
              )}
            </>
          ) : (
            <p>
              Game over! Last word: {currentWord?.scrambled} → <span className="font-bold">{lastWordAnswer}</span>
            </p>
          ))}
      </div>

      <div className="flex flex-col gap-2 mb-3 w-full text-center">
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
          className="col-span-4 py-2 md:py-3 px-4 text-sm font-medium text-dark-text-secondary border border-dark-border-primary rounded transition duration-200 hover:bg-dark-bg-tertiary hover:text-dark-text-primary flex gap-1 items-center justify-center"
          data-ga-label="ga-back-to-home-button"
        >
          {isMobile ? (
            <>
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="m15 18-6-6 6-6" />
              </svg>

              <span>Home</span>
            </>
          ) : (
            'Back to home'
          )}
        </button>
        <div className="col-span-6 flex">
          <SettingsModal
            triggerText={'⚙️'}
            triggerClassnames="border-r-0 rounded-r-none py-2 md:py-3"
            settings={settings}
            updateSettings={onUpdateSettings}
          />

          <button
            type="button"
            onClick={onPlayAgain}
            className="flex-1 py-2 md:py-3 px-4 text-sm font-semibold bg-dark-interactive-primary text-white rounded rounded-l-none hover:shadow-lg active:translate-y-0 shadow hover:bg-dark-interactive-hover"
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

function highlightAnswer(definition: string, answer: string): ReactNode {
  const parts = definition.split(answer);
  return parts.map((part, i) => (
    <Fragment key={i}>
      {part}
      {i < parts.length - 1 && <strong className="underline">{answer}</strong>}
    </Fragment>
  ));
}
