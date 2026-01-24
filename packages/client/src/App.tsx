import { useGame } from './hooks/useGame';
import { StartScreen } from './components/StartScreen';
import { GameScreen } from './components/GameScreen';
import { ResultsScreen } from './components/ResultsScreen';
import './styles.css';

export default function App() {
  const { state, startGame, submitAnswer, resetGame } = useGame();

  if (state.phase === 'idle') {
    return <StartScreen onStart={startGame} />;
  }

  if (state.phase === 'playing' && state.currentWord && state.scrambled) {
    return (
      <GameScreen
        scrambled={state.scrambled}
        definition={state.currentWord.definition}
        score={state.score}
        totalAttempts={state.totalAttempts}
        remainingSeconds={state.remainingSeconds}
        expectedWord={state.currentWord.word}
        autoVoice={state.autoVoice}
        onSubmit={submitAnswer}
      />
    );
  }

  if (state.phase === 'finished') {
    return (
      <ResultsScreen
        score={state.score}
        totalAttempts={state.totalAttempts}
        leaderboard={state.gameResults}
        onPlayAgain={resetGame}
      />
    );
  }

  return null;
}
