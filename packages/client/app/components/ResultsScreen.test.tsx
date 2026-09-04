import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useGame, useLeaderboard, useLeaderboardPercentile } from '~/api/hooks';
import { readStoredSettings } from '~/lib/settings';

import { ResultsScreen } from './ResultsScreen';

vi.mock('~/api/hooks', async (importOriginal) => ({
  ...((await importOriginal()) as any),
  useLeaderboardPercentile: vi.fn(() => ({ data: undefined })),
  useLeaderboard: vi.fn(() => ({ data: undefined, isFetching: false, error: null })),
}));

interface RenderOverrides {
  onPlayAgain?: () => void;
  onBackToHome?: () => void;
  score?: number;
  totalAttempts?: number;
  currentWord?: { scrambled: string; definition: string } | null;
  lastWordAnswer?: string | null;
  lastWordDefinition?: string | null;
  lastWordReferences?: string[] | null;
  correctAttemptTimestamps?: string[][];
}

function renderScreen(overrides: RenderOverrides = {}) {
  const client = new QueryClient();

  function Wrapper() {
    const { state, updateSettings } = useGame(false, readStoredSettings());

    return (
      <ResultsScreen
        score={overrides.score ?? 3}
        totalAttempts={overrides.totalAttempts ?? 5}
        currentWord={overrides.currentWord ?? { scrambled: 'plepa', definition: 'A thin, flat cake.' }}
        lastWordAnswer={overrides.lastWordAnswer ?? 'apple'}
        lastWordDefinition={overrides.lastWordDefinition ?? null}
        lastWordReferences={overrides.lastWordReferences ?? null}
        correctAttemptTimestamps={overrides.correctAttemptTimestamps ?? []}
        onPlayAgain={overrides.onPlayAgain ?? vi.fn()}
        onBackToHome={overrides.onBackToHome ?? vi.fn()}
        settings={state.settings}
        onUpdateSettings={updateSettings}
      />
    );
  }

  return render(
    <QueryClientProvider client={client}>
      <Wrapper />
    </QueryClientProvider>,
  );
}

describe('ResultsScreen', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it('calls onPlayAgain when Play Again is clicked', async () => {
    const user = userEvent.setup();
    const onPlayAgain = vi.fn();
    renderScreen({ onPlayAgain });

    await user.click(screen.getByRole('button', { name: 'Play Again' }));

    expect(onPlayAgain).toHaveBeenCalledTimes(1);
  });

  it('calls onBackToHome when Back to home is clicked', async () => {
    const user = userEvent.setup();
    const onBackToHome = vi.fn();
    renderScreen({ onBackToHome });

    await user.click(screen.getByRole('button', { name: 'Back to home' }));

    expect(onBackToHome).toHaveBeenCalledTimes(1);
  });

  it('renders a Settings trigger next to Play Again', () => {
    renderScreen();

    expect(screen.getByRole('button', { name: 'Settings' })).toBeInTheDocument();
  });

  it('persists the auto-voice toggle across re-renders', async () => {
    const user = userEvent.setup();
    const { unmount } = renderScreen();

    await user.click(screen.getByRole('button', { name: 'Settings' }));
    await user.click(screen.getByRole('checkbox', { name: /Disabled/ }));

    unmount();
    renderScreen();

    await user.click(screen.getByRole('button', { name: 'Settings' }));

    expect(screen.getByRole('checkbox', { name: /Enabled/ })).toBeChecked();
  });

  it('prefixes the last word line with "Game over!" for the english-words topic', () => {
    renderScreen();

    expect(screen.getByText(/Game over! Last word:/)).toBeInTheDocument();
  });

  it('shows "Last quote:" with bolded answer for indonesian-politician-quotes topic', () => {
    localStorage.setItem('atoyr:settings:v2', JSON.stringify({ autoVoice: false, mode: 'vanilla', topic: 'indonesian-politician-quotes' }));
    const rawDef = 'Kalau ada yang bilang itu Indonesia <template>, yang <template> kau, bukan Indonesia';
    renderScreen({
      lastWordAnswer: 'gelap',
      lastWordDefinition: rawDef.replace(/<template>/g, 'gelap'),
      currentWord: { scrambled: 'lgepa', definition: rawDef },
    });

    expect(screen.getByText(/Game over! Last quote:/)).toBeInTheDocument();
    const strongElements = screen.getAllByText('gelap');
    expect(strongElements).toHaveLength(2);
    strongElements.forEach((el) => {
      expect(el.tagName).toBe('STRONG');
    });
  });

  it('renders an All-time / This month button group in the leaderboard section', () => {
    renderScreen();

    const radiogroup = screen.getByRole('radiogroup', { name: 'Leaderboard period' });
    expect(within(radiogroup).getByRole('radio', { name: 'All-time' })).toHaveAttribute('aria-checked', 'true');
    expect(within(radiogroup).getByRole('radio', { name: 'This month' })).toHaveAttribute('aria-checked', 'false');
  });

  it('switches the leaderboard and percentile queries to monthly when This month is selected', async () => {
    const user = userEvent.setup();
    renderScreen();

    await user.click(screen.getByRole('radio', { name: 'This month' }));

    expect(useLeaderboard).toHaveBeenCalledWith(expect.any(Object), 'monthly', undefined, 5);
    expect(useLeaderboardPercentile).toHaveBeenCalledWith(expect.any(Object), 'monthly');
  });

  it('shows percentile above the leaderboard', () => {
    vi.mocked(useLeaderboardPercentile).mockReturnValue({ data: { percentile: 75, rank: 3 } } as any);
    vi.mocked(useLeaderboard).mockReturnValue({
      data: {
        entries: [
          {
            id: 'current-session',
            isSessionSameAsCurrentUser: true,
            topic: 'english-words',
            score: 10,
            totalAttempts: 5,
            accuracy: 80,
            timestamp: 0,
            rank: 3,
          },
        ],
        total: 5,
      },
      isFetching: false,
      error: null,
    } as any);

    renderScreen();

    expect(
      screen.getByText(
        (_, element) => element?.tagName === 'P' && element?.textContent.includes('Your result was better than 75% of players!'),
      ),
    ).toBeInTheDocument();
  });
});
