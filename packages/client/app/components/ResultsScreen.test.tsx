import { screen, render, cleanup } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { ResultsScreen } from './ResultsScreen';

vi.mock('~/api/hooks', () => ({
  useLeaderboardPercentile: () => ({ data: undefined }),
  useLeaderboard: () => ({ data: undefined, isFetching: false, error: null }),
}));

interface RenderOverrides {
  onPlayAgain?: () => void;
  onBackToHome?: () => void;
  score?: number;
  totalAttempts?: number;
  currentWord?: { scrambled: string; definition: string } | null;
  lastWordAnswer?: string | null;
  correctAttemptTimestamps?: string[][];
}

function renderScreen(overrides: RenderOverrides = {}) {
  return render(
    <ResultsScreen
      score={overrides.score ?? 3}
      totalAttempts={overrides.totalAttempts ?? 5}
      currentWord={overrides.currentWord ?? { scrambled: 'plepa', definition: 'A thin, flat cake.' }}
      lastWordAnswer={overrides.lastWordAnswer ?? 'apple'}
      correctAttemptTimestamps={overrides.correctAttemptTimestamps ?? []}
      onPlayAgain={overrides.onPlayAgain ?? vi.fn()}
      onBackToHome={overrides.onBackToHome ?? vi.fn()}
    />,
  );
}

describe('ResultsScreen', () => {
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
});
