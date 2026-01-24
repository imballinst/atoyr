import React from 'react';

interface StartScreenProps {
  onStart: () => void;
}

export function StartScreen({ onStart }: StartScreenProps) {
  return (
    <div className="start-screen">
      <div className="start-card">
        <h1>Atoyr</h1>
        <p className="subtitle">A Test of Your Reflexes</p>

        <div className="instructions">
          <p>Unscramble 5-letter words as fast as you can!</p>
          <ul>
            <li>You have 30 seconds</li>
            <li>Wrong answers cost 1 second</li>
            <li>Type using the keyboard or click the on-screen buttons</li>
            <li>Press Enter or click Submit to guess</li>
          </ul>
        </div>

        <button className="start-btn" onClick={onStart}>
          Start Game
        </button>
      </div>
    </div>
  );
}
