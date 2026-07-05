import React, { useCallback, useEffect, useState } from 'react';

import type { GameSessionState } from '~/lib/game';

import { Keyboard } from './Keyboard';

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
  const [feedback, setFeedback] = useState<'correct' | 'incorrect' | null>(null);

  const speakLetters = (letters: string, definition: string) => {
    if (!('speechSynthesis' in window)) return;
    window.speechSynthesis.cancel();

    const definitionUtterance = new SpeechSynthesisUtterance(definition);
    definitionUtterance.rate = 0.75;
    window.speechSynthesis.speak(definitionUtterance);

    letters.split('').forEach((letter) => {
      const utterance = new SpeechSynthesisUtterance(letter);
      window.speechSynthesis.speak(utterance);
    });
  };

  const handleLetterClick = (letter: string) => {
    setAnswer((prev) => {
      if (prev.length === 5) return prev;
      return answer + letter;
    });
  };

  const handleBackspace = () => {
    setAnswer((prev) => prev.slice(0, -1));
  };

  const handleSubmit = useCallback(() => {
    if (answer.trim().length === 0) return;
    setFeedback(null);
    onSubmit(answer, token, {
      onSuccess: () => {
        setFeedback('correct');
        setTimeout(() => setFeedback(null), 1000);
      },
      onError: () => {
        setFeedback('incorrect');
        setTimeout(() => setFeedback(null), 1000);
      },
    });
    setAnswer('');
  }, [answer, token]);

  useEffect(() => {
    // Small delay to ensure focus is set before speaking
    const speakTimeout = setTimeout(() => {
      if (autoVoice) {
        speakLetters(scrambled, definition);
      }
    }, 50);
    return () => clearTimeout(speakTimeout);
  }, [scrambled, definition, autoVoice]);

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === 'Enter') {
        return handleSubmit();
      } else if (e.key === 'Backspace') {
        return handleBackspace();
      }

      const lowerCased = e.key.toLowerCase();
      if (/[a-z]/.test(lowerCased)) {
        handleLetterClick(lowerCased);
      }
    }

    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [handleSubmit]);

  const accuracy = totalAttempts > 0 ? ((score / totalAttempts) * 100).toFixed(1) : '0.0';
  const currentStreak = correctAttemptTimestamps[correctAttemptTimestamps.length - 1] ?? [];

  return (
    <div className="w-full h-full flex flex-col gap-4 justify-end items-center">
      <div className="w-full flex flex-col gap-2">
        <div
          className={`text-4xl font-mono font-bold bg-dark-bg-tertiary p-3 rounded-lg text-center ${remainingSeconds <= 5 ? 'text-red-500' : 'text-dark-text-primary'} w-full`}
        >
          {remainingSeconds}s
        </div>
        <div className="flex gap-2 flex-1 w-full">
          <div className="flex-1 bg-dark-bg-tertiary p-3 rounded-lg text-center">
            <div className="text-xs text-dark-text-tertiary">
              {score}/{totalAttempts} correct
            </div>
            <div className="text-sm font-semibold text-dark-text-primary">{accuracy}%</div>
          </div>
          <div className="flex-1 bg-dark-bg-tertiary p-3 rounded-lg text-center">
            <div className="text-xs text-dark-text-tertiary">Streak</div>
            <div className="text-sm font-semibold text-dark-text-primary">{currentStreak.length}</div>
          </div>
        </div>
      </div>

      <div className="border border-dark-bg-tertiary p-4 rounded-lg text-center text-sm italic text-dark-text-secondary min-h-10 flex items-center justify-center w-full">
        {definition}
      </div>

      <div className="flex gap-2 justify-center w-full">
        {autoVoice ? (
          <div className="sr-only" role="status" aria-live="polite" aria-label={`Letters to unscramble: ${scrambled}`}>
            {scrambled.split('').map((letter, i) => (
              <span key={i}>{letter}</span>
            ))}
          </div>
        ) : (
          scrambled.split('').map((letter, i) => (
            <div
              key={i}
              className="w-12 h-12 flex items-center justify-center bg-dark-interactive-primary text-white font-bold text-2xl rounded-lg shadow uppercase"
            >
              {letter}
            </div>
          ))
        )}
      </div>

      <button
        onClick={() => speakLetters(scrambled, definition)}
        className="bg-dark-interactive-primary text-white w-12 h-12 rounded-full text-2xl transition duration-200 hover:bg-dark-interactive-hover hover:scale-110 active:scale-95"
        aria-label="Speak letters"
      >
        🔊
      </button>

      <div className="flex gap-2 justify-center w-full">
        {answer.split('').map((letter, i) => (
          <div
            key={i}
            className="w-12 h-12 flex items-center justify-center bg-dark-bg-tertiary border-2 border-dark-border-primary font-bold text-2xl rounded-lg text-dark-text-primary"
          >
            {letter.toUpperCase()}
          </div>
        ))}
        {answer.length < 5 &&
          Array.from({ length: 5 - answer.length }).map((_, i) => (
            <div key={`empty-${i}`} className="w-12 h-12 bg-dark-bg-accent border-2 border-dark-border-primary rounded-lg" />
          ))}
      </div>

      <Keyboard answer={answer} onBackspace={handleBackspace} onClick={handleLetterClick} onSubmit={handleSubmit} />

      {feedback && (
        <div
          className={`fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-8xl animate-fade-in-out z-50 ${
            feedback === 'correct' ? 'text-green-500' : 'text-red-500'
          }`}
        >
          {feedback === 'correct' ? '✓' : '✗'}
        </div>
      )}
    </div>
  );
}
