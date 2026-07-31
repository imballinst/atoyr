import { nanoid } from 'nanoid';
import { Fragment, type ReactNode, useEffect, useRef, useState } from 'react';

import type { GameState } from '~/lib/game';

import { Keyboard } from './Keyboard';

const DEFAULT_LANG = 'en-US';
const TEMPLATE_PRONUNCIATION: Record<string, string> = {
  [DEFAULT_LANG]: 'dot dot dot',
  'id-ID': 'titik titik titik',
};

interface GameScreenProps extends GameState {
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
  settings,
  token,
  lang,
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

  const submitIfComplete = (nextAnswer: string, token: string, wordLen: number) => {
    if (nextAnswer.trim().length !== wordLen) return;

    onSubmit(nextAnswer, token, {
      onSuccess: () => showFeedback(true),
      onError: () => showFeedback(false),
    });
    answerRef.current = '';
    setAnswer('');
  };

  const handleLetterClick = (letter: string, token: string, wordLen: number) => {
    if (answerRef.current.length >= wordLen) return;

    const next = answerRef.current + letter;
    answerRef.current = next;
    setAnswer(next);
    submitIfComplete(next, token, wordLen);
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
      if (/^[a-z]$/.test(lowerCased)) handleLetterClick(lowerCased, token, scrambled.length);
    }

    window.addEventListener('keydown', onKeyDown);
    return () => {
      window.removeEventListener('keydown', onKeyDown);
      window.speechSynthesis.cancel();
    };
    // handleLetterClick/handleBackspace read answerRef.current (always latest);
    // submitIfComplete closes over onSubmit/token from this render.
    // Re-bind only when onSubmit/token change (rare: phase change / new word).
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [onSubmit, token]);

  useEffect(() => {
    const speakTimeout = setTimeout(() => {
      if (settings.autoVoice) {
        speakLetters(scrambled, definition, lang);
      }
    }, 50);
    return () => clearTimeout(speakTimeout);
  }, [scrambled, definition, settings.autoVoice, lang]);

  const accuracy = totalAttempts > 0 ? ((score / totalAttempts) * 100).toFixed(1) : '0.0';
  const currentStreak = correctAttemptTimestamps[correctAttemptTimestamps.length - 1] ?? [];
  const definitionContent = renderDefinitionVisual(definition, scrambled.length);
  const isIndonesianTopic = settings.topic === 'indonesian-politician-quotes';

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

      <div
        className={
          'border border-dark-bg-tertiary p-4 rounded-lg justify-center items-center flex text-sm italic text-dark-text-secondary w-full' +
          (isIndonesianTopic ? ' h-[114px] overflow-hidden line-clamp-4' : '')
        }
        hidden={settings.mode === 'blind' || settings.autoVoice}
      >
        {definitionContent}
      </div>

      <div
        className="flex gap-2 justify-center w-full"
        role="status"
        aria-live="polite"
        aria-atomic="true"
        aria-label={`Letters to unscramble: ${scrambled}`}
      >
        {settings.autoVoice ? (
          <div className="sr-only">
            {scrambled.split('').map((letter, i) => (
              <span key={i}>{letter}</span>
            ))}
          </div>
        ) : (
          Array.from({ length: scrambled.length }, (_, i) => (
            <div
              key={i}
              className="w-12 h-12 flex items-center justify-center bg-dark-interactive-primary text-white font-bold text-2xl rounded-lg shadow uppercase"
            >
              <div className="sr-only">{ordinalLabel(i)} char: </div>
              {scrambled[i]}
            </div>
          ))
        )}
      </div>

      <button
        type="button"
        onClick={() => speakLetters(scrambled, definition, lang)}
        className="bg-dark-interactive-primary text-white w-12 h-12 rounded-full text-2xl transition duration-200 hover:bg-dark-interactive-hover hover:scale-110 active:scale-95"
        aria-label="Speak letters"
      >
        🔊
      </button>

      <div className="flex gap-2 justify-center w-full" data-testid="answer-slots">
        {Array.from({ length: scrambled.length }, (_, i) => {
          if (i < answer.length) {
            return (
              <div
                key={i}
                className="w-12 h-12 flex items-center justify-center bg-dark-bg-tertiary border-2 border-dark-border-primary font-bold text-2xl rounded-lg text-dark-text-primary"
              >
                <div className="sr-only">{ordinalLabel(i)} char: </div>
                {answer[i].toUpperCase()}
              </div>
            );
          }
          return <div key={i} className="w-12 h-12 bg-dark-bg-accent border-2 border-dark-border-primary rounded-lg" />;
        })}
      </div>

      <Keyboard
        answer={answer}
        wordLength={scrambled.length}
        onBackspace={handleBackspace}
        onClick={(letter) => handleLetterClick(letter, token, scrambled.length)}
      />

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

function makeUtterance(text: string, lang: string) {
  const utterance = new SpeechSynthesisUtterance(text);
  if (lang !== DEFAULT_LANG) utterance.lang = lang;
  utterance.volume = 0.75;
  return utterance;
}

function speakLetters(letters: string, definition: string, lang: string) {
  if (!('speechSynthesis' in window)) return;
  window.speechSynthesis.cancel();

  if (definition) {
    const speechDefinition = definition.replace(/<template>(-<template>)?/g, TEMPLATE_PRONUNCIATION[lang] ?? '...');
    const definitionUtterance = makeUtterance(speechDefinition, lang);
    definitionUtterance.rate = 0.75;
    window.speechSynthesis.speak(definitionUtterance);
  }

  letters.split('').forEach((letter) => {
    window.speechSynthesis.speak(makeUtterance(letter, lang));
  });
}

function ordinalLabel(i: number): string {
  const labels = [
    'First',
    'Second',
    'Third',
    'Fourth',
    'Fifth',
    'Sixth',
    'Seventh',
    'Eighth',
    'Ninth',
    'Tenth',
    'Eleventh',
    'Twelfth',
    'Thirteenth',
    'Fourteenth',
    'Fifteenth',
    'Sixteenth',
    'Seventeenth',
    'Eighteenth',
    'Nineteenth',
    'Twentieth',
  ];
  return i < labels.length ? labels[i] : `${i + 1}th`;
}

function renderDefinitionVisual(definition: string, scrambledLength: number): ReactNode {
  const template = '<template>';
  if (!definition.includes(template)) return definition;

  const underscores = '_'.repeat(scrambledLength);
  return definition.replace(/<template>/g, underscores);
}

function getClassNames() {
  const translates = ['translate-y-10', 'translate-y-20', 'translate-y-30'];
  return Array.from({ length: 3 }, () => translates[Math.floor(Math.random() * translates.length)]) as [string, string, string];
}
