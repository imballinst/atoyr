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

  useEffect(() => {
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
    if (answer.length < 5) {
      setAnswer(answer + letter);
    }
  };

  const handleBackspace = () => {
    setAnswer(answer.slice(0, -1));
  };

  const handleSubmit = () => {
    if (answer.trim().length === 0) return;
    onSubmit(answer);
    setFeedback(null);
    setAnswer('');
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
  const timerColor = remainingSeconds <= 5 ? 'text-red-500' : 'text-gray-800';

  return (
    <div className="w-screen h-screen max-w-2xl mx-auto flex flex-col items-center justify-center p-4 bg-gray-50 overflow-hidden">
      <div className="w-full max-w-[430px] mx-auto flex justify-between items-center mb-6 gap-4">
        <div className={`text-4xl font-bold min-w-20 text-center bg-white p-3 rounded-lg shadow ${timerColor}`}>
          {remainingSeconds}s
        </div>
        <div className="flex gap-4 flex-1">
          <div className="flex-1 bg-white p-3 rounded-lg text-center shadow">
            <div className="text-xs text-gray-500">{score} correct</div>
            <div className="text-sm font-semibold text-gray-800">{accuracy}%</div>
          </div>
          <div className="flex-1 bg-white p-3 rounded-lg text-center shadow">
            <div className="text-xs text-gray-500">attempts</div>
            <div className="text-sm font-semibold text-gray-800">{totalAttempts}</div>
          </div>
        </div>
      </div>

      <div className="flex flex-col items-center w-full max-w-[430px] mx-auto gap-4">
        <div className="bg-white p-4 rounded-lg text-center text-sm italic text-gray-500 min-h-10 flex items-center justify-center shadow w-full">
          {definition}
        </div>

        <div className="flex gap-2 justify-center w-full">
          {scrambled.split('').map((letter, i) => (
            <div
              key={i}
              className="w-12 h-12 flex items-center justify-center bg-gradient-to-br from-indigo-500 to-purple-600 text-white font-bold text-2xl rounded-lg shadow"
            >
              {letter}
            </div>
          ))}
        </div>

        <button
          onClick={() => speakLetters(scrambled)}
          className="bg-blue-500 text-white w-12 h-12 rounded-full text-2xl transition duration-200 hover:bg-blue-700 hover:scale-110 active:scale-95 shadow"
          aria-label="Speak letters"
        >
          🔊
        </button>

        <div className="flex gap-2 justify-center w-full">
          {answer.split('').map((letter, i) => (
            <div key={i} className="w-12 h-12 flex items-center justify-center bg-white border-2 border-gray-300 font-bold text-2xl rounded-lg shadow">
              {letter}
            </div>
          ))}
          {answer.length < 5 &&
            Array.from({ length: 5 - answer.length }).map((_, i) => (
              <div key={`empty-${i}`} className="w-12 h-12 bg-gray-100 border-2 border-gray-300 rounded-lg shadow" />
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
          className="absolute opacity-0 pointer-events-none"
          aria-label="Answer input"
        />

        <div className="w-full max-w-[430px] mx-auto flex flex-col gap-3">
          <div className="grid grid-cols-10 gap-1.5 w-full">
            {['Q', 'W', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P'].map((key) => (
              <button
                key={key}
                className="aspect-square bg-white border-2 border-gray-300 rounded font-semibold text-xs transition duration-200 shadow hover:bg-blue-500 hover:text-white hover:-translate-y-0.5 hover:shadow-lg active:translate-y-0 disabled:opacity-50 disabled:cursor-not-allowed"
                onClick={() => handleLetterClick(key)}
                disabled={answer.length >= 5}
              >
                {key}
              </button>
            ))}
          </div>
          <div className="grid grid-cols-9 gap-1.5 w-full">
            {['A', 'S', 'D', 'F', 'G', 'H', 'J', 'K', 'L'].map((key) => (
              <button
                key={key}
                className="aspect-square bg-white border-2 border-gray-300 rounded font-semibold text-xs transition duration-200 shadow hover:bg-blue-500 hover:text-white hover:-translate-y-0.5 hover:shadow-lg active:translate-y-0 disabled:opacity-50 disabled:cursor-not-allowed"
                onClick={() => handleLetterClick(key)}
                disabled={answer.length >= 5}
              >
                {key}
              </button>
            ))}
          </div>
          <div className="grid grid-cols-7 gap-1.5 w-full">
            {['Z', 'X', 'C', 'V', 'B', 'N', 'M'].map((key) => (
              <button
                key={key}
                className="aspect-square bg-white border-2 border-gray-300 rounded font-semibold text-xs transition duration-200 shadow hover:bg-blue-500 hover:text-white hover:-translate-y-0.5 hover:shadow-lg active:translate-y-0 disabled:opacity-50 disabled:cursor-not-allowed"
                onClick={() => handleLetterClick(key)}
                disabled={answer.length >= 5}
              >
                {key}
              </button>
            ))}
          </div>
        </div>

        <div className="grid grid-cols-3 gap-1.5 w-full">
          <button
            onClick={handleBackspace}
            className="col-span-1 py-3 bg-red-500 text-white font-semibold text-sm rounded transition duration-200 shadow hover:bg-red-600 hover:-translate-y-0.5 hover:shadow-lg active:translate-y-0"
            aria-label="Backspace"
          >
            ← Back
          </button>
          <div className="col-span-2" />
          <button
            onClick={handleSubmit}
            disabled={answer.length === 0}
            className="col-span-3 py-3 bg-green-500 text-white font-semibold text-sm rounded transition duration-200 shadow hover:bg-green-600 hover:-translate-y-0.5 hover:shadow-lg active:translate-y-0 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Submit
          </button>
        </div>

        {feedback && (
          <div className={`fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-8xl animate-fade-in-out z-50 ${feedback === 'correct' ? 'text-green-500' : 'text-red-500'
            }`}>
            {feedback === 'correct' ? '✓' : '✗'}
          </div>
        )}
      </div>
    </div>
  );
}
