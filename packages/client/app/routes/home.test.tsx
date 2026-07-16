import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { type ReactNode } from 'react';
import { createRoutesStub } from 'react-router';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { apiStartGame, apiSubmitAnswer } from '~/api/client';
import type { StartGameResponse, SubmitAnswerResponse } from '~/api/gen';
import { LATEST_STORAGE_KEY, readStoredSettings, type LatestSchema } from '~/lib/settings';
import Home from '~/routes/home';

const sseRef = vi.hoisted(() => ({
  onEvent: (_data: Record<string, unknown>) => {},
  onError: (_error: Error) => {},
}));

vi.mock('~/api/client', () => ({
  apiStartGame: vi.fn(),
  apiResumeGame: vi.fn(),
  apiSubmitAnswer: vi.fn(),
  apiSubscribeToSSE: vi.fn((onEvent: (d: Record<string, unknown>) => void, onError: (e: Error) => void) => {
    sseRef.onEvent = onEvent;
    sseRef.onError = onError;
    return vi.fn();
  }),
  apiQuery: { useQuery: vi.fn(() => ({ data: undefined })) },
}));

const mockStartResponse1: StartGameResponse = {
  sessionId: 'session-1',
  mode: 'vanilla',
  remainingSeconds: 30,
  scrambledWord: 'plepa',
  scrambledWordDefinition: 'A thin, flat cake made from batter.',
  token: 'token-1',
  autoVoice: false,
};

const mockStartResponse2: StartGameResponse = {
  sessionId: 'session-2',
  mode: 'vanilla',
  remainingSeconds: 30,
  scrambledWord: 'rhcea',
  scrambledWordDefinition: 'A sweet baked food.',
  token: 'token-2',
  autoVoice: false,
};

describe('Home — game lifecycle', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
    sseRef.onEvent = (_data: Record<string, unknown>) => {};
    sseRef.onError = (_error: Error) => {};
    mockSpeechSynthesis();
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it('transitions from finished to playing with a new word and resets score/timer', async () => {
    vi.mocked(apiStartGame).mockResolvedValueOnce(mockStartResponse1).mockResolvedValueOnce(mockStartResponse2);
    renderHome();

    await startGame();
    await finishCurrentGame();

    await userEvent.click(screen.getByRole('button', { name: 'Play Again' }));
    await waitFor(() => expect(screen.getByRole('heading', { name: 'Remaining seconds' })).toBeInTheDocument());
  });

  it('transitions from finished to idle when Back to home is clicked', async () => {
    vi.mocked(apiStartGame).mockResolvedValue(mockStartResponse1);
    renderHome();

    await startGame();
    await finishCurrentGame();

    await userEvent.click(screen.getByRole('button', { name: 'Back to home' }));

    expect(screen.getByRole('button', { name: 'Start Game' })).toBeInTheDocument();
  });

  it('preserves auto-voice preference from localStorage when playing again', async () => {
    localStorage.setItem(LATEST_STORAGE_KEY, JSON.stringify({ autoVoice: true, mode: 'vanilla' } satisfies LatestSchema));
    vi.mocked(apiStartGame)
      .mockResolvedValueOnce({ ...mockStartResponse1, autoVoice: true })
      .mockResolvedValueOnce({ ...mockStartResponse2, autoVoice: true });
    renderHome();

    await startGame();
    await finishCurrentGame();

    await userEvent.click(screen.getByRole('button', { name: 'Play Again' }));
    await waitFor(() => expect(screen.getByRole('heading', { name: 'Remaining seconds' })).toBeInTheDocument());

    expect(apiStartGame).toHaveBeenCalledTimes(2);
    expect(apiStartGame).toHaveBeenNthCalledWith(1, { autoVoice: true, mode: 'vanilla' }, []);
    expect(apiStartGame).toHaveBeenNthCalledWith(2, { autoVoice: true, mode: 'vanilla' }, []);
  });

  it('keeps app on results screen when replay start fails', async () => {
    const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {});
    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    vi.mocked(apiStartGame).mockResolvedValueOnce(mockStartResponse1).mockRejectedValueOnce(new Error('Server error'));
    renderHome();

    await startGame();
    await finishCurrentGame();

    await userEvent.click(screen.getByRole('button', { name: 'Play Again' }));
    await waitFor(() => expect(alertSpy).toHaveBeenCalled());

    expect(screen.getByRole('heading', { name: 'Game Over!' })).toBeInTheDocument();
    alertSpy.mockRestore();
    consoleErrorSpy.mockRestore();
  });

  it.each([
    { mode: 'vanilla' as const, wordDefinition: mockStartResponse1.scrambledWordDefinition, bannerText: '' as const },
    { mode: 'blind' as const, wordDefinition: '' as const, bannerText: '🚫 BLIND MODE 🚫' as const },
  ])('shows feedback and updates the scrambled word after a correct answer in $mode mode', async ({ mode, wordDefinition, bannerText }) => {
    localStorage.setItem(LATEST_STORAGE_KEY, JSON.stringify({ autoVoice: false, mode } satisfies LatestSchema));
    vi.mocked(apiStartGame).mockResolvedValue({
      ...mockStartResponse1,
      mode,
      scrambledWordDefinition: wordDefinition,
    });
    vi.mocked(apiSubmitAnswer).mockResolvedValue({
      correct: true,
      scrambledWord: 'rahce',
      scrambledWordDefinition: wordDefinition,
      token: 'token-2',
      score: 1,
      attempts: 1,
      remainingSeconds: 28,
      correctAttemptTimestamps: [['2024-01-01T00:00:00Z']],
    } satisfies SubmitAnswerResponse);
    renderHome();

    await startGame();

    if (bannerText) expect(screen.getByText(bannerText)).toBeInTheDocument();

    for (const letter of ['P', 'L', 'E', 'P', 'A']) {
      await userEvent.click(screen.getByRole('button', { name: letter }));
    }

    await waitFor(() => expect(screen.getAllByText('🎉')).toHaveLength(3));

    expect(screen.getByRole('status', { name: 'Letters to unscramble: rahce' })).toBeInTheDocument();
  });

  it('uses the updated persisted auto-voice value when replaying after toggling in settings', async () => {
    vi.mocked(apiStartGame)
      .mockResolvedValueOnce(mockStartResponse1)
      .mockResolvedValueOnce({ ...mockStartResponse2, autoVoice: true });
    renderHome();

    await startGame();
    await finishCurrentGame();

    await userEvent.click(screen.getByRole('button', { name: 'Settings' }));
    await userEvent.click(screen.getByRole('checkbox', { name: /Disabled/ }));
    await userEvent.click(screen.getByRole('button', { name: 'Close' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());

    await userEvent.click(screen.getByRole('button', { name: 'Play Again' }));
    await waitFor(() => expect(screen.getByRole('heading', { name: 'Remaining seconds' })).toBeInTheDocument());

    expect(apiStartGame).toHaveBeenCalledTimes(2);
    expect(apiStartGame).toHaveBeenNthCalledWith(1, { autoVoice: false, mode: 'vanilla' }, []);
    expect(apiStartGame).toHaveBeenNthCalledWith(2, { autoVoice: true, mode: 'vanilla' }, []);
  });
});

function mockSpeechSynthesis() {
  const speak = vi.fn();
  const cancel = vi.fn();
  const synthesis = { speak, cancel, getVoices: () => [] } as unknown as SpeechSynthesis;

  Object.defineProperty(window, 'speechSynthesis', {
    value: synthesis,
    writable: true,
    configurable: true,
  });

  Object.defineProperty(window, 'SpeechSynthesisUtterance', {
    value: class {
      text: string;
      rate = 1;
      constructor(text: string) {
        this.text = text;
      }
    },
    writable: true,
    configurable: true,
  });

  return { speak, cancel };
}

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

function renderHome(shouldFetch = false) {
  const RemixStub = createRoutesStub([
    {
      id: 'home',
      path: '/',
      Component: Home,
      loader() {
        return { shouldFetch, settings: readStoredSettings() };
      },
    },
  ]);

  return render(
    <RemixStub initialEntries={['/']} hydrationData={{ loaderData: { home: { shouldFetch, settings: readStoredSettings() } } }} />,
    {
      wrapper: createWrapper(),
    },
  );
}

async function startGame() {
  await userEvent.click(screen.getByRole('button', { name: 'Start Game' }));
  await waitFor(() => expect(screen.getByRole('heading', { name: 'Remaining seconds' })).toBeInTheDocument());
}

async function finishCurrentGame() {
  await waitFor(() => expect(sseRef.onEvent).not.toBeNull());
  sseRef.onEvent({ type: 'finish', lastWordAnswer: 'apple' });
  await waitFor(() => expect(screen.getByRole('heading', { name: 'Game Over!' })).toBeInTheDocument());
}
