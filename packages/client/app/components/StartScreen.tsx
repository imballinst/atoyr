import React, { useState } from 'react';

interface StartScreenProps {
  onStart: (autoVoice: boolean) => void;
}

export function StartScreen({ onStart }: StartScreenProps) {
  const [autoVoice, setAutoVoice] = useState(() => {
    if (typeof window === 'undefined') return false;

    const stored = localStorage.getItem('atoyr_auto_voice');
    if (stored === null) return false;

    const value = JSON.parse(stored);
    return typeof value === 'boolean' ? value : false;
  });

  const handleAutoVoiceChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.checked;
    setAutoVoice(newValue);
    localStorage.setItem('atoyr_auto_voice', JSON.stringify(newValue));
  };

  const handleStart = () => {
    onStart(autoVoice);
  };

  return (
    <div className="rounded-2xl text-center w-full">
      <h1 className="text-5xl font-bold text-dark-interactive-primary mb-3">Atoyr</h1>
      <p className="text-base text-dark-text-secondary font-medium mb-10">A Test of Your Reflexes</p>

      <div className="text-left mb-10">
        <p className="text-sm text-dark-text-secondary mb-6">Unscramble 5-letter words as fast as you can!</p>
        <ul className="text-xs text-dark-text-tertiary space-y-2">
          <li className="flex gap-3">
            <span className="text-dark-interactive-success font-bold shrink-0">✓</span>
            <span>You have 30 seconds</span>
          </li>
          <li className="flex gap-3">
            <span className="text-dark-interactive-success font-bold shrink-0">✓</span>
            <span>Wrong answers cost 1 second</span>
          </li>
          <li className="flex gap-3">
            <span className="text-dark-interactive-success font-bold shrink-0">✓</span>
            <span>Type using the keyboard or click the on-screen buttons</span>
          </li>
          <li className="flex gap-3">
            <span className="text-dark-interactive-success font-bold shrink-0">✓</span>
            <span>Press Enter or click Submit to guess</span>
          </li>
        </ul>
      </div>

      <div className="flex items-center gap-3 mb-4">
        <input type="checkbox" id="auto-voice" checked={autoVoice} onChange={handleAutoVoiceChange} className="w-4 h-4 cursor-pointer" />
        <label htmlFor="auto-voice" className="text-sm text-left text-dark-text-secondary cursor-pointer flex-1">
          Enable automatic text-to-speech*
        </label>
      </div>

      <button
        onClick={handleStart}
        className="w-full py-3 px-6 text-base font-semibold bg-dark-interactive-primary text-white rounded-lg transition duration-200 hover:shadow-lg active:translate-y-0 shadow hover:bg-dark-interactive-hover"
      >
        Start Game
      </button>

      <hr className="my-8 border-t border-t-dark-border-primary" />

      <p className="text-xs text-dark-text-tertiary italic text-left">
        * Uses your browser's built-in text-to-speech functionality. You will get 5 extra seconds for each word, but the scrambled letters
        will only be shown for screen readers.
      </p>
    </div>
  );
}
