import React from 'react';

interface StartScreenProps {
  onStart: () => void;
}

export function StartScreen({ onStart }: StartScreenProps) {
  return (
    <div className="w-screen h-screen max-w-2xl mx-auto flex items-center justify-center p-5 bg-dark-bg-primary">
      <div className="rounded-2xl p-10 text-center w-full max-w-[430px]">
        <h1 className="text-5xl font-bold text-dark-interactive-primary mb-6">
          Atoyr
        </h1>
        <p className="text-base text-dark-text-secondary font-medium mb-10">A Test of Your Reflexes</p>

        <div className="text-left mb-10">
          <p className="text-sm text-dark-text-secondary mb-6">Unscramble 5-letter words as fast as you can!</p>
          <ul className="text-xs text-dark-text-tertiary space-y-3">
            <li className="flex gap-3">
              <span className="text-dark-interactive-success font-bold flex-shrink-0">✓</span>
              <span>You have 30 seconds</span>
            </li>
            <li className="flex gap-3">
              <span className="text-dark-interactive-success font-bold flex-shrink-0">✓</span>
              <span>Wrong answers cost 1 second</span>
            </li>
            <li className="flex gap-3">
              <span className="text-dark-interactive-success font-bold flex-shrink-0">✓</span>
              <span>Type using the keyboard or click the on-screen buttons</span>
            </li>
            <li className="flex gap-3">
              <span className="text-dark-interactive-success font-bold flex-shrink-0">✓</span>
              <span>Press Enter or click Submit to guess</span>
            </li>
          </ul>
        </div>

        <button
          onClick={onStart}
          className="w-full py-3 px-6 text-base font-semibold bg-dark-interactive-primary text-white rounded-lg transition duration-200 hover:-translate-y-0.5 hover:shadow-lg active:translate-y-0 shadow hover:bg-dark-interactive-hover"
        >
          Start Game
        </button>
      </div>
    </div>
  );
}
