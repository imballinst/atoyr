import { GameScreen } from './components/GameScreen';
import { ResultsScreen } from './components/ResultsScreen';
import { StartScreen } from './components/StartScreen';
import { useServerGame } from './hooks/useServerGame';
import './styles.css';

export default function App() {
  const gameHook = useServerGame();
  const { state, startGame, submitAnswer, resetGame } = gameHook;

  if (state.phase === 'idle') {
    return <StartScreen onStart={startGame} />;
  }

  const { currentWord, currentWordToken } = state;

  if (state.phase === 'playing' && currentWord && currentWordToken) {
    return (
      <GameScreen
        token={currentWordToken}
        scrambled={currentWord.scrambled}
        definition={currentWord.definition}
        score={state.score}
        totalAttempts={state.totalAttempts}
        correctAttemptTimestamps={state.correctAttemptTimestamps}
        remainingSeconds={state.remainingSeconds}
        autoVoice={state.autoVoice}
        onSubmit={submitAnswer}
      />
    );
  }

  if (state.phase === 'finished') {
    return (
      <ResultsScreen score={state.score} totalAttempts={state.totalAttempts} leaderboard={state.gameResults} onPlayAgain={resetGame} />
    );
  }

  return null;
}
