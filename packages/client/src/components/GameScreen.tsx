import React, { useEffect, useRef, useState } from 'react';

interface GameScreenProps {
  scrambled: string;
  definition: string;
  score: number;
  totalAttempts: number;
  remainingSeconds: number;
  onSubmit: (answer: string) => void;
}

export function GameScreen({
  scrambled,
  definition,
  score,
  totalAttempts,
  remainingSeconds,
  onSubmit,
}: GameScreenProps) {
  const [answer, setAnswer] = useState('');
  const [feedback, setFeedback] = useState<'correct' | 'incorrect' | null>(null);
  const textInputRef = useRef<HTMLInputElement>(null);
  const spelledOutRef = useRef<string[]>([]);

  useEffect(() => {
    spelledOutRef.current = [];
    if (textInputRef.current) {
      textInputRef.current.focus();
    }
  }, [scrambled]);

  const speakLetters = (letters: string) => {
    if (!('speechSynthesis' in window)) return;
    window.speechSynthesis.cancel();
    letters.split('').forEach((letter, i) => {
      const utterance = new SpeechSynthesisUtterance(letter);
      utterance.rate = 0.8;
      setTimeout(() => window.speechSynthesis.speak(utterance), i * 400);
    });
  };

  const handleLetterClick = (letter: string) => {
    const newAnswer = answer + letter;
    setAnswer(newAnswer);
    spelledOutRef.current.push(letter);
  };

  const handleBackspace = () => {
    if (answer.length > 0) {
      const newAnswer = answer.slice(0, -1);
      setAnswer(newAnswer);
      spelledOutRef.current.pop();
    }
  };

  const handleSubmit = () => {
    if (answer.trim().length === 0) return;
    onSubmit(answer);
    setFeedback(null);
    setAnswer('');
    spelledOutRef.current = [];
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSubmit();
    } else if (e.key === 'Backspace') {
      e.preventDefault();
      handleBackspace();
    }
  };

  const accuracy = totalAttempts > 0 ? ((score / totalAttempts) * 100).toFixed(1) : '0.0';

  return (
    <div className="game-screen">
      <div className="game-header">
        <div className="timer" style={{ color: remainingSeconds <= 5 ? '#ff4444' : '#333' }}>
          {remainingSeconds}s
        </div>
        <div className="score-info">
          <div className="score">{score} correct</div>
          <div className="accuracy">{accuracy}% accuracy</div>
        </div>
      </div>

      <div className="game-content">
        <div className="definition-box">{definition}</div>

        <div className="scrambled-display">
          {scrambled.split('').map((letter, i) => (
            <div key={i} className="letter-box">
              {letter}
            </div>
          ))}
        </div>

        <button className="speaker-btn" onClick={() => speakLetters(scrambled)} aria-label="Speak letters">
          🔊
        </button>

        <div className="answer-display">
          {answer.split('').map((letter, i) => (
            <div key={i} className="answer-box">
              {letter}
            </div>
          ))}
          {answer.length < 5 &&
            Array.from({ length: 5 - answer.length }).map((_, i) => (
              <div key={`empty-${i}`} className="answer-box empty" />
            ))}
        </div>

        <input
          ref={textInputRef}
          type="text"
          value={answer}
          onChange={(e) => setAnswer(e.target.value.toUpperCase())}
          onKeyDown={handleKeyDown}
          maxLength={5}
          autoComplete="off"
          className="hidden-input"
          aria-label="Answer input"
        />

        <div className="keyboard-container">
          <div className="keyboard">
            {['Q', 'W', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P'].map((key) => (
              <button
                key={key}
                className="key"
                onClick={() => handleLetterClick(key)}
                disabled={answer.length >= 5}
              >
                {key}
              </button>
            ))}
            {['A', 'S', 'D', 'F', 'G', 'H', 'J', 'K', 'L'].map((key) => (
              <button
                key={key}
                className="key"
                onClick={() => handleLetterClick(key)}
                disabled={answer.length >= 5}
              >
                {key}
              </button>
            ))}
            {['Z', 'X', 'C', 'V', 'B', 'N', 'M'].map((key) => (
              <button
                key={key}
                className="key"
                onClick={() => handleLetterClick(key)}
                disabled={answer.length >= 5}
              >
                {key}
              </button>
            ))}
          </div>
          <div className="keyboard-actions">
            <button className="action-btn backspace" onClick={handleBackspace} aria-label="Backspace">
              ← Back
            </button>
            <button className="action-btn submit" onClick={handleSubmit} disabled={answer.length === 0}>
              Submit
            </button>
          </div>
        </div>

        {feedback && <div className={`feedback ${feedback}`}>{feedback === 'correct' ? '✓' : '✗'}</div>}
      </div>
    </div>
  );
}
