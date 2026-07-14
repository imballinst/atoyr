import { nanoid } from 'nanoid';
import { Fragment, useEffect, useRef, useState } from 'react';

import type { GameSessionState } from '~/lib/game';

import { Keyboard } from './Keyboard';

const ORDINAL_LABELS = ['First', 'Second', 'Third', 'Fourth', 'Fifth'];

interface GameScreenProps extends GameSessionState {
  scrambled: string;
  definition: string;
  token: string;
  onSubmit: (answer: string, token: string, { onSuccess, onError }: { onSuccess?: () => void; onError?: () => void }) => void;
}

export function GameScreen({
  scrambled,
  definition,
  score,
  totalAttempts,
  correctAttemptTimestamps,
  remainingSeconds,
  autoVoice,
  token,
  onSubmit,
}: GameScreenProps) {
  const [answer, setAnswer] = useState('');
  const [feedback, setFeedback] = useState<Array<{ id: string; isCorrect: boolean; classNames: [string, string, string] }>>([]);

  const answerRef = useRef(answer);

  const showFeedback = (isCorrect: boolean) => {
    const id = nanoid();

    setFeedback((prev) => prev.concat({ id, isCorrect, classNames: getClassNames() }));
    setTimeout(() => setFeedback((prev) => prev.filter((item) => item.id !== id)), 1500);
  };

  const submitIfComplete = (nextAnswer: string, token: string) => {
    if (nextAnswer.trim().length !== 5) return;

    onSubmit(nextAnswer, token, {
      onSuccess: () => showFeedback(true),
      onError: () => showFeedback(false),
    });
    answerRef.current = '';
    setAnswer('');
  };

  const handleLetterClick = (letter: string, token: string) => {
    if (answerRef.current.length >= 5) return;

    const next = answerRef.current + letter;
    answerRef.current = next;
    setAnswer(next);
    submitIfComplete(next, token);
  };

  const handleBackspace = () => {
    const next = answerRef.current.slice(0, -1);
    answerRef.current = next;
    setAnswer(next);
  };

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === 'Backspace') return handleBackspace();

      const lowerCased = e.key.toLowerCase();
      if (/^[a-z]$/.test(lowerCased)) handleLetterClick(lowerCased, token);
    }

    window.addEventListener('keydown', onKeyDown);
    return () => {
      window.removeEventListener('keydown', onKeyDown);
    };
    // handleLetterClick/handleBackspace read answerRef.current (always latest);
    // submitIfComplete closes over onSubmit/token from this render.
    // Re-bind only when onSubmit/token change (rare: phase change / new word).
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [onSubmit, token]);

  useEffect(() => {
    const speakTimeout = setTimeout(() => {
      if (autoVoice) {
        speakLetters(scrambled, definition);
      }
    }, 50);
    return () => clearTimeout(speakTimeout);
  }, [scrambled, definition, autoVoice]);

  const accuracy = totalAttempts > 0 ? ((score / totalAttempts) * 100).toFixed(1) : '0.0';
  const currentStreak = correctAttemptTimestamps[correctAttemptTimestamps.length - 1] ?? [];

  return (
    <div className="w-full h-full flex flex-col gap-4 justify-end items-center">
      <div className="w-full flex flex-col gap-2">
        <section
          className={`text-4xl font-mono font-bold bg-dark-bg-tertiary p-3 rounded-lg text-center ${remainingSeconds <= 5 ? 'text-red-500' : 'text-dark-text-primary'} w-full`}
        >
          <h3 className="sr-only">Remaining seconds</h3>
          {remainingSeconds}s
        </section>
        <div className="flex gap-2 flex-1 w-full">
          <section className="flex-1 bg-dark-bg-tertiary p-3 rounded-lg text-center">
            <h3 className="sr-only">Score</h3>
            <div className="text-xs text-dark-text-tertiary">
              {score}/{totalAttempts} correct
            </div>
            <div className="text-sm font-semibold text-dark-text-primary">{accuracy}%</div>
          </section>
          <section className="flex-1 bg-dark-bg-tertiary p-3 rounded-lg text-center">
            <h3 className="text-xs text-dark-text-tertiary">Streak</h3>
            <div className="text-sm font-semibold text-dark-text-primary">{currentStreak.length}</div>
          </section>
        </div>
      </div>

      <div className="border border-dark-bg-tertiary p-4 rounded-lg text-center text-sm italic text-dark-text-secondary min-h-10 flex items-center justify-center w-full">
        {definition}
      </div>

      <div
        className="flex gap-2 justify-center w-full"
        role="status"
        aria-live="polite"
        aria-atomic="true"
        aria-label={`Letters to unscramble: ${scrambled}`}
      >
        {autoVoice ? (
          <div className="sr-only">
            {ORDINAL_LABELS.map((label, i) => (
              <span key={label}>{scrambled[i]}</span>
            ))}
          </div>
        ) : (
          ORDINAL_LABELS.map((label, i) => (
            <div
              key={label}
              className="w-12 h-12 flex items-center justify-center bg-dark-interactive-primary text-white font-bold text-2xl rounded-lg shadow uppercase"
            >
              <div className="sr-only">{label} char: </div>
              {scrambled[i]}
            </div>
          ))
        )}
      </div>

      <button
        type="button"
        onClick={() => speakLetters(scrambled, definition)}
        className="bg-dark-interactive-primary text-white w-12 h-12 rounded-full text-2xl transition duration-200 hover:bg-dark-interactive-hover hover:scale-110 active:scale-95"
        aria-label="Speak letters"
      >
        🔊
      </button>

      <div className="flex gap-2 justify-center w-full" data-testid="answer-slots">
        {ORDINAL_LABELS.slice(0, answer.length).map((label, i) => (
          <div
            key={label}
            className="w-12 h-12 flex items-center justify-center bg-dark-bg-tertiary border-2 border-dark-border-primary font-bold text-2xl rounded-lg text-dark-text-primary"
          >
            <div className="sr-only">{label} char: </div>
            {answer[i].toUpperCase()}
          </div>
        ))}
        {answer.length < 5 &&
          ORDINAL_LABELS.slice(answer.length).map((label) => (
            <div key={label} className="w-12 h-12 bg-dark-bg-accent border-2 border-dark-border-primary rounded-lg" />
          ))}
      </div>

      <Keyboard answer={answer} onBackspace={handleBackspace} onClick={(letter) => handleLetterClick(letter, token)} />

      {feedback.map(({ id, isCorrect, classNames }) => {
        const className =
          'absolute px-4 py-2 rounded-full bg-black text-white text-sm font-medium pointer-events-none animate-float-up-fade';
        const rendered = isCorrect ? '🎉' : '❌';

        return (
          <Fragment key={id}>
            <div className={'-translate-x-40 ' + className + ` ${classNames[0]}`}>{rendered}</div>
            <div className={className + ` ${classNames[1]}`}>{rendered}</div>
            <div className={'translate-x-40 ' + className + ` ${classNames[2]}`}>{rendered}</div>
          </Fragment>
        );
      })}
    </div>
  );
}

function speakLetters(letters: string, definition: string) {
  if (!('speechSynthesis' in window)) return;
  window.speechSynthesis.cancel();

  const definitionUtterance = new SpeechSynthesisUtterance(definition);
  definitionUtterance.rate = 0.75;
  window.speechSynthesis.speak(definitionUtterance);

  letters.split('').forEach((letter) => {
    const utterance = new SpeechSynthesisUtterance(letter);
    window.speechSynthesis.speak(utterance);
  });
}

function getClassNames() {
  const translates = ['translate-y-10', 'translate-y-20', 'translate-y-30'];
  return Array.from({ length: 3 }, () => translates[Math.floor(Math.random() * translates.length)]) as [string, string, string];
}
