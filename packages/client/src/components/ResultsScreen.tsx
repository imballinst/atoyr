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
    <div className="results-screen">
      <div className="results-card">
        <h1>Game Over!</h1>

        <div className="final-stats">
          <div className="stat">
            <div className="stat-label">Correct</div>
            <div className="stat-value">{score}</div>
          </div>
          <div className="stat">
            <div className="stat-label">Total Attempts</div>
            <div className="stat-value">{totalAttempts}</div>
          </div>
          <div className="stat">
            <div className="stat-label">Accuracy</div>
            <div className="stat-value">{accuracy}%</div>
          </div>
        </div>

        <button className="play-again-btn" onClick={onPlayAgain}>
          Play Again
        </button>

        {leaderboard.length > 0 && (
          <div className="leaderboard">
            <h2>Leaderboard</h2>
            <div className="leaderboard-list">
              {leaderboard
                .sort((a, b) => b.score - a.score)
                .slice(0, 10)
                .map((result, i) => (
                  <div key={result.id} className="leaderboard-entry">
                    <span className="rank">#{i + 1}</span>
                    <span className="entry-score">{result.score}</span>
                    <span className="entry-accuracy">{(result.accuracy * 100).toFixed(0)}%</span>
                  </div>
                ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
