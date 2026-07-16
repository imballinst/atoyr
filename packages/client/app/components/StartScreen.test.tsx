import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useGame } from '~/api/hooks';
import { readStoredSettings, writeStoredSettings } from '~/lib/settings';

import { StartScreen } from './StartScreen';

function renderScreen(overrides: { onStart?: () => void } = {}) {
  const client = new QueryClient();

  function Wrapper() {
    const { state, updateSettings } = useGame(false, readStoredSettings());

    return <StartScreen onStart={overrides.onStart ?? vi.fn()} settings={state.settings} updateSettings={updateSettings} />;
  }

  return render(
    <QueryClientProvider client={client}>
      <Wrapper />
    </QueryClientProvider>,
  );
}

async function openHowToPlay(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole('button', { name: 'How to play' }));
}

async function openSettings(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole('button', { name: 'Settings' }));
}

describe('StartScreen', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it('does not render the rules list inline by default', () => {
    renderScreen();

    expect(screen.queryByRole('heading', { name: 'How to play' })).not.toBeInTheDocument();
  });

  it('does not render the auto-voice checkbox or footnote inline by default', () => {
    renderScreen();

    expect(screen.queryByRole('heading', { name: 'Settings' })).not.toBeInTheDocument();
  });

  it('opens the How to play modal when clicked', async () => {
    const user = userEvent.setup();
    renderScreen();

    await openHowToPlay(user);

    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'How to play' })).toBeInTheDocument();
  });

  it('opens the Settings modal and shows the auto-voice toggle', async () => {
    const user = userEvent.setup();
    renderScreen();

    await openSettings(user);

    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Settings' })).toBeInTheDocument();
    expect(screen.getByRole('checkbox', { name: /Disabled/ })).toBeInTheDocument();
  });

  it('uses the persisted auto-voice value in the settings modal', async () => {
    writeStoredSettings({ autoVoice: true, mode: 'vanilla' });
    const user = userEvent.setup();
    renderScreen();

    await openSettings(user);

    expect(screen.getByRole('checkbox', { name: /Enabled/ })).toBeChecked();
  });

  it('persists the auto-voice toggle across re-renders', async () => {
    const user = userEvent.setup();
    const { unmount } = renderScreen();

    await openSettings(user);
    await user.click(screen.getByRole('checkbox', { name: /Disabled/ }));

    unmount();
    renderScreen();

    await openSettings(user);

    expect(screen.getByRole('checkbox', { name: /Enabled/ })).toBeChecked();
  });

  it('starts the game when the Start Game button is clicked', async () => {
    const user = userEvent.setup();
    const onStart = vi.fn();
    renderScreen({ onStart });

    await user.click(screen.getByRole('button', { name: 'Start Game' }));

    expect(onStart).toHaveBeenCalledTimes(1);
  });
});
