import React from 'react';
import { GameResult } from '../types/game';

interface ResultsScreenProps {
  score: number;
  totalAttempts: number;
  leaderboard: GameResult[];
  onPlayAgain: () => void;
}

export function ResultsScreen({ score, totalAttempts, leaderboard, onPlayAgain }: ResultsScreenProps) {
  const accuracy = totalAttempts > 0 ? ((score / totalAttempts) * 100).toFixed(1) : '0.0';

  return (
    <div className="w-screen h-screen max-w-2xl mx-auto flex items-center justify-center p-5 bg-gradient-to-br from-blue-400 via-indigo-400 to-purple-500">
      <div className="bg-white rounded-2xl p-10 text-center shadow-2xl w-full max-w-sm">
        <h1 className="text-4xl font-bold mb-6 text-gray-800">Game Over!</h1>

        <div className="grid grid-cols-3 gap-3 mb-8">
          <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
            <div className="text-xs text-gray-500 mb-2 font-medium">Correct</div>
            <div className="text-3xl font-bold text-blue-500">{score}</div>
          </div>
          <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
            <div className="text-xs text-gray-500 mb-2 font-medium">Attempts</div>
            <div className="text-3xl font-bold text-gray-800">{totalAttempts}</div>
          </div>
          <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
            <div className="text-xs text-gray-500 mb-2 font-medium">Accuracy</div>
            <div className="text-3xl font-bold text-green-500">{accuracy}%</div>
          </div>
        </div>

        <button
          onClick={onPlayAgain}
          className="w-full py-3 px-6 text-base font-semibold bg-gradient-to-r from-indigo-500 to-purple-600 text-white rounded-lg transition duration-200 hover:-translate-y-0.5 hover:shadow-lg active:translate-y-0 shadow mb-6"
        >
          Play Again
        </button>

        {leaderboard.length > 0 && (
          <div className="mt-6 pt-6 border-t border-gray-300">
            <h2 className="text-lg font-semibold mb-4 text-gray-800">Leaderboard</h2>
            <div className="flex flex-col gap-2 max-h-52 overflow-y-auto">
              {leaderboard
                .sort((a, b) => b.score - a.score)
                .slice(0, 10)
                .map((result, i) => (
                  <div key={result.id} className="flex gap-3 p-3 bg-gray-50 rounded text-xs">
                    <span className="font-semibold text-blue-500 min-w-8">#{i + 1}</span>
                    <span className="flex-1 text-center font-semibold text-gray-800">{result.score}</span>
                    <span className="font-semibold text-green-500 min-w-12">
                      {(result.accuracy * 100).toFixed(0)}%
                    </span>
                  </div>
                ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
