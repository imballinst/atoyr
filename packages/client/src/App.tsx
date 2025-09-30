import React, { useState } from 'react';
import { pickRandomWord } from '@atoyr/shared';

export default function App() {
  const [words, setWords] = useState<string[]>(['apple', 'banana', 'cherry', 'date', 'elder']);
  const [log, setLog] = useState<string[]>([]);

  function startMatch() {
    const myWord = pickRandomWord(words);
    setLog((l) => [`Picked word: ${myWord}`, ...l]);
  }

  return (
    <div className="app">
      <h1>A Toy R — Prototype</h1>
      <div>
        <label>My words (5):</label>
        <ul>
          {words.map((w) => (
            <li key={w}>{w}</li>
          ))}
        </ul>
      </div>
      <button onClick={startMatch}>Start Mock Match</button>
      <div className="log">
        {log.map((l, i) => (
          <div key={i}>{l}</div>
        ))}
      </div>
    </div>
  );
}
