import { GameScreen } from '~/components/GameScreen';
import { ResultsScreen } from '~/components/ResultsScreen';
import { StartScreen } from '~/components/StartScreen';
import { useGame } from '~/api/hooks';
import type { Route } from './+types/home';

export function meta({ }: Route.MetaArgs) {
  return [{ title: 'Game | atoyr' }, { name: 'description', content: 'Welcome to React Router!' }];
}

export default function Home() {
  const { state, startGame, submitAnswer, resetGame } = useGame();

  if (state.phase === 'idle') {
    return <StartScreen onStart={startGame} />;
  }

  const { currentWord, currentWordToken } = state;

  if (state.phase === 'playing' && currentWord && currentWordToken) {
    return (
      <GameScreen
        {...state}
        token={currentWordToken}
        definition={currentWord.definition}
        scrambled={currentWord.scrambled}
        onSubmit={submitAnswer}
      />
    );
  }

  if (state.phase === 'finished') {
    return (
      <ResultsScreen
        score={state.score}
        totalAttempts={state.totalAttempts}
        onPlayAgain={resetGame}
        correctAttemptTimestamps={state.correctAttemptTimestamps}
      />
    );
  }

  return null;
}
