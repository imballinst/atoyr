import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useGame } from '~/api/hooks';
import { encodeSettings } from '~/lib/settings';

import { SettingsModal } from './SettingsModal';

function renderModal() {
  const client = new QueryClient();

  function Wrapper() {
    const { state, updateSettings } = useGame(false, { autoVoice: false, mode: 'blind', topic: 'english-words' });

    return <SettingsModal settings={state.settings} updateSettings={updateSettings} />;
  }

  return render(
    <QueryClientProvider client={client}>
      <Wrapper />
    </QueryClientProvider>,
  );
}

describe('SettingsModal — share settings', () => {
  let writeText: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    localStorage.clear();
    writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText },
      writable: true,
      configurable: true,
    });
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it('copies a share URL to the clipboard and shows Copied! feedback', async () => {
    renderModal();

    await userEvent.click(screen.getByRole('button', { name: 'Settings' }));

    expect(screen.getByRole('dialog')).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Share settings' }));

    await waitFor(() =>
      expect(writeText).toHaveBeenCalledWith(
        `${location.origin}/?settings=${encodeSettings({ autoVoice: false, mode: 'blind', topic: 'english-words' })}`,
      ),
    );

    expect(screen.getByRole('button', { name: 'Copied!' })).toBeInTheDocument();

    await waitFor(() => expect(screen.getByRole('button', { name: 'Share settings' })).toBeInTheDocument(), { timeout: 2500 });
  });
});
