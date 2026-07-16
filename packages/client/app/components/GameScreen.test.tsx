import { fireEvent, screen, render, waitFor, cleanup, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { GameState } from '~/lib/game';

import { GameScreen } from './GameScreen';

type SubmitFn = (answer: string, token: string, callbacks: { onSuccess?: () => void; onError?: () => void }) => void;

interface RenderOverrides extends Partial<GameState> {
  scrambled?: string;
  definition?: string;
  token?: string;
  onSubmit?: SubmitFn;
}

const DEFAULT_SCRAMBLED = 'plepa';
const DEFAULT_DEFINITION = 'A thin, flat cake made from batter.';
const DEFAULT_TOKEN = 'token-1';

const baseState: GameState = {
  phase: 'playing',
  score: 2,
  totalAttempts: 4,
  remainingSeconds: 20,
  usedWords: [],
  currentWordToken: 'token-1',
  correctAttemptTimestamps: [
    ['2024-01-01T00:00:00Z', '2024-01-01T00:00:05Z'],
    ['2024-01-01T00:00:10Z', '2024-01-01T00:00:15Z'],
  ],
  currentWord: {
    definition: 'example',
    scrambled: 'example',
  },
  lastWordAnswer: null,
  settings: {
    autoVoice: false,
    mode: 'vanilla',
  },
};

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

function renderScreen(overrides: RenderOverrides = {}) {
  const { onSubmit = vi.fn(), scrambled, definition, token, ...stateOverrides } = overrides;

  return render(
    <GameScreen
      {...baseState}
      {...stateOverrides}
      scrambled={scrambled ?? DEFAULT_SCRAMBLED}
      definition={definition ?? DEFAULT_DEFINITION}
      token={token ?? DEFAULT_TOKEN}
      onSubmit={onSubmit}
    />,
  );
}

function emptyAnswerSlotCount(container: HTMLElement): number {
  return container.querySelectorAll('div.bg-dark-bg-accent').length;
}

function scrambledTileCount(container: HTMLElement): number {
  return container.querySelectorAll('div.bg-dark-interactive-primary').length;
}

// Matches an answer slot whose text is split across an sr-only label + a visible letter.
// Constraint: the slot is a direct child of [data-testid="answer-slots"], disambiguating
// from the container whose textContent also aggregates the slot's value.
function answerSlotMatcher(text: string) {
  return (_content: string, node: Element | null) =>
    !!node && node.parentElement?.dataset.testid === 'answer-slots' && node.textContent === text;
}

describe('GameScreen', () => {
  beforeEach(() => {
    mockSpeechSynthesis();
  });

  afterEach(() => {
    cleanup();
  });

  it('renders the definition, timer, score and accuracy', () => {
    renderScreen();

    expect(screen.getByText('A thin, flat cake made from batter.')).toBeInTheDocument();
    expect(within(screen.getByRole('heading', { name: 'Remaining seconds' }).closest('section')!).getByText('20s')).toBeInTheDocument();
    const scoreSection = screen.getByRole('heading', { name: 'Score' }).closest('section')!;
    expect(within(scoreSection).getByText('2/4 correct')).toBeInTheDocument();
    expect(within(scoreSection).getByText('50.0%')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Streak' })).toBeInTheDocument();
  });

  it('renders 5 scrambled letter tiles when autoVoice is off', () => {
    const { container } = renderScreen({ scrambled: 'plepa' });

    expect(scrambledTileCount(container)).toBe(5);
  });

  it('announces the scrambled letters via a live status region', () => {
    renderScreen({ scrambled: 'plepa' });

    expect(screen.getByRole('status', { name: 'Letters to unscramble: plepa' })).toBeInTheDocument();
  });

  it('renders 5 empty answer slots initially', () => {
    const { container } = renderScreen();

    expect(emptyAnswerSlotCount(container)).toBe(5);
  });

  it('appends a letter when a keyboard key is clicked', async () => {
    const user = userEvent.setup();
    const { container } = renderScreen();

    await user.click(screen.getByRole('button', { name: 'P' }));

    expect(emptyAnswerSlotCount(container)).toBe(4);
    const answerSlots = screen.getByTestId('answer-slots');
    expect(within(answerSlots).getByText(answerSlotMatcher('First char: P'))).toBeInTheDocument();
  });

  it('removes the last letter when backspace is clicked', async () => {
    const user = userEvent.setup();
    const { container } = renderScreen();

    await user.click(screen.getByRole('button', { name: 'P' }));
    await user.click(screen.getByRole('button', { name: 'Backspace' }));

    expect(emptyAnswerSlotCount(container)).toBe(5);
    const answerSlots = screen.getByTestId('answer-slots');
    expect(within(answerSlots).queryByText(answerSlotMatcher('First char: P'))).not.toBeInTheDocument();
  });

  it('does not submit before 5 letters are typed', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn<SubmitFn>();
    const { container } = renderScreen({ onSubmit });

    for (const letter of ['P', 'L', 'E']) {
      await user.click(screen.getByRole('button', { name: letter }));
    }

    expect(onSubmit).not.toHaveBeenCalled();
    expect(emptyAnswerSlotCount(container)).toBe(2);
    const answerSlots = screen.getByTestId('answer-slots');
    expect(within(answerSlots).getByText(answerSlotMatcher('First char: P'))).toBeInTheDocument();
    expect(within(answerSlots).getByText(answerSlotMatcher('Second char: L'))).toBeInTheDocument();
    expect(within(answerSlots).getByText(answerSlotMatcher('Third char: E'))).toBeInTheDocument();
  });

  it('auto-submits when 5 letters are typed and shows success feedback on onSuccess', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn<SubmitFn>();
    onSubmit.mockImplementation((_answer, _token, { onSuccess }) => onSuccess?.());
    renderScreen({ onSubmit });

    for (const letter of ['P', 'L', 'E', 'P', 'A']) {
      await user.click(screen.getByRole('button', { name: letter }));
    }

    expect(onSubmit).toHaveBeenCalledTimes(1);
    expect(onSubmit).toHaveBeenCalledWith('PLEPA', 'token-1', expect.objectContaining({ onSuccess: expect.any(Function) }));

    expect(await screen.findAllByText('🎉')).toHaveLength(3);

    await waitFor(() => expect(screen.queryAllByText('🎉')).toHaveLength(0), { timeout: 3000 });
    const answerSlots = screen.getByTestId('answer-slots');
    expect(within(answerSlots).queryByText(answerSlotMatcher('First char: P'))).not.toBeInTheDocument();
  });

  it('shows error feedback when onError is invoked', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn<SubmitFn>();
    onSubmit.mockImplementation((_answer, _token, { onError }) => onError?.());
    renderScreen({ onSubmit });

    for (const letter of ['P', 'L', 'E', 'P', 'A']) {
      await user.click(screen.getByRole('button', { name: letter }));
    }

    expect(await screen.findAllByText('❌')).toHaveLength(3);

    await waitFor(() => expect(screen.queryAllByText('❌')).toHaveLength(0), { timeout: 3000 });
  });

  it('clears the answer after auto-submitting', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn<SubmitFn>();
    onSubmit.mockImplementation((_answer, _token, { onSuccess }) => onSuccess?.());
    const { container } = renderScreen({ onSubmit });

    for (const letter of ['P', 'L', 'E', 'P', 'A']) {
      await user.click(screen.getByRole('button', { name: letter }));
    }

    expect(emptyAnswerSlotCount(container)).toBe(5);
    const answerSlots = screen.getByTestId('answer-slots');
    expect(within(answerSlots).queryByText(answerSlotMatcher('First char: P'))).not.toBeInTheDocument();
  });

  it('accepts new letters after an auto-submitted answer (answerRef is cleared)', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn<SubmitFn>();
    onSubmit.mockImplementation((_answer, _token, { onSuccess }) => onSuccess?.());
    const { container } = renderScreen({ onSubmit });

    for (const letter of ['P', 'L', 'E', 'P', 'A']) {
      await user.click(screen.getByRole('button', { name: letter }));
    }
    expect(onSubmit).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole('button', { name: 'P' }));

    expect(emptyAnswerSlotCount(container)).toBe(4);
    const answerSlots = screen.getByTestId('answer-slots');
    expect(within(answerSlots).getByText(answerSlotMatcher('First char: P'))).toBeInTheDocument();
  });

  it('speaks letters and definition when the speak button is clicked', async () => {
    const user = userEvent.setup();
    const { cancel, speak } = mockSpeechSynthesis();
    renderScreen();

    await user.click(screen.getByRole('button', { name: 'Speak letters' }));

    expect(cancel).toHaveBeenCalled();
    expect(speak).toHaveBeenCalled();
    expect((speak.mock.calls[0][0] as { text: string }).text).toBe('A thin, flat cake made from batter.');
    expect((speak.mock.calls[1][0] as { text: string }).text).toBe('p');
  });

  it('does not auto-speak on mount when autoVoice is off', () => {
    mockSpeechSynthesis();
    renderScreen({ settings: { autoVoice: false, mode: 'vanilla' } });

    expect(window.speechSynthesis.speak).not.toHaveBeenCalled();
  });

  it('auto-speaks on mount when autoVoice is on', async () => {
    const { speak } = mockSpeechSynthesis();
    renderScreen({ settings: { autoVoice: true, mode: 'vanilla' } });

    await waitFor(() => expect(speak).toHaveBeenCalled());
  });

  it('records typed letters when physical keyboard keys are pressed', () => {
    const { container } = renderScreen();

    fireEvent.keyDown(window, { key: 'p' });
    fireEvent.keyDown(window, { key: 'l' });

    expect(emptyAnswerSlotCount(container)).toBe(3);
    const answerSlots = screen.getByTestId('answer-slots');
    expect(within(answerSlots).getByText(answerSlotMatcher('First char: P'))).toBeInTheDocument();
    expect(within(answerSlots).getByText(answerSlotMatcher('Second char: L'))).toBeInTheDocument();
  });

  it('ignores non-letter keys on the physical keyboard', () => {
    const { container } = renderScreen();

    fireEvent.keyDown(window, { key: '1' });
    fireEvent.keyDown(window, { key: ' ' });

    expect(emptyAnswerSlotCount(container)).toBe(5);
  });

  it('handles backspace via the physical keyboard', () => {
    const { container } = renderScreen();

    fireEvent.keyDown(window, { key: 'p' });
    fireEvent.keyDown(window, { key: 'Backspace' });

    expect(emptyAnswerSlotCount(container)).toBe(5);
    const answerSlots = screen.getByTestId('answer-slots');
    expect(within(answerSlots).queryByText(answerSlotMatcher('First char: P'))).not.toBeInTheDocument();
  });

  it('keeps the streak display in sync with the last correct attempt group', () => {
    renderScreen({
      correctAttemptTimestamps: [['2024-01-01T00:00:00Z'], ['2024-01-01T00:00:10Z', '2024-01-01T00:00:15Z', '2024-01-01T00:00:20Z']],
    });

    const streakSection = screen.getByRole('heading', { name: 'Streak' }).closest('section')!;
    expect(within(streakSection).getByText('3')).toBeInTheDocument();
  });

  it('hides the definition in blind mode', () => {
    renderScreen({ settings: { autoVoice: false, mode: 'blind' }, definition: DEFAULT_DEFINITION });

    expect(screen.getByText(DEFAULT_DEFINITION)).toHaveAttribute('hidden');
  });

  it('speaks only letters when the definition is empty', async () => {
    const user = userEvent.setup();
    const { cancel, speak } = mockSpeechSynthesis();
    renderScreen({ definition: '' });

    await user.click(screen.getByRole('button', { name: 'Speak letters' }));

    expect(cancel).toHaveBeenCalled();
    expect(speak.mock.calls.map((c) => (c[0] as { text: string }).text)).toEqual(['p', 'l', 'e', 'p', 'a']);
  });
});
