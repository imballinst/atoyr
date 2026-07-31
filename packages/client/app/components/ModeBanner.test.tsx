import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';

import { ModeBanner } from './ModeBanner';

describe('ModeBanner', () => {
  afterEach(() => {
    cleanup();
  });

  it.each([
    { mode: 'vanilla' as const, topic: 'english-words' as const, expected: '🍦 Vanilla mode + English Words 🍦' },
    { mode: 'blind' as const, topic: 'english-words' as const, expected: '🚫 BLIND MODE + English Words 🚫' },
    {
      mode: 'vanilla' as const,
      topic: 'indonesian-politician-quotes' as const,
      expected: '🍦 Vanilla mode + Indonesian Politician Quotes 🍦',
    },
    { mode: 'blind' as const, topic: 'indonesian-politician-quotes' as const, expected: '🚫 BLIND MODE + Indonesian Politician Quotes 🚫' },
  ])('renders topic label: $topic with mode: $mode', ({ mode, topic, expected }) => {
    render(<ModeBanner mode={mode} topic={topic} />);

    const banner = screen.getByRole('status');
    expect(banner.textContent).toBe(expected);
  });

  it('renders with correct aria-label for each topic', () => {
    render(<ModeBanner mode="vanilla" topic="indonesian-politician-quotes" />);

    expect(screen.getByRole('status')).toHaveAttribute('aria-label', 'Vanilla mode, Indonesian Politician Quotes topic');
  });
});
