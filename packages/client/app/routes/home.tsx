import { Loader2Icon } from 'lucide-react';
import { useLoaderData } from 'react-router';

import { useGame } from '~/api/hooks';
import { GameScreen } from '~/components/GameScreen';
import { ResultsScreen } from '~/components/ResultsScreen';
import { StartScreen } from '~/components/StartScreen';
import { hasGameEnded } from '~/lib/game';

export function meta() {
  return [{ title: 'Game | atoyr' }, { name: 'description', content: 'Welcome to React Router!' }];
}

export function clientLoader() {
  return { shouldFetch: hasGameEnded() };
}

export default function Home() {
  const { shouldFetch } = useLoaderData<typeof clientLoader>();
  const { state, startGame, submitAnswer, resetGame, resumeGameQuery } = useGame(shouldFetch);

  if (resumeGameQuery.isFetching) {
    return (
      <div className="w-full h-full text-dark-text-primary flex flex-col items-center justify-center gap-y-2">
        <Loader2Icon className="animate-spin" />
        <div>Resuming your game...</div>
      </div>
    );
  }

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
        currentWord={state.currentWord}
        lastWordAnswer={state.lastWordAnswer}
        onPlayAgain={resetGame}
        correctAttemptTimestamps={state.correctAttemptTimestamps}
      />
    );
  }

  return null;
}
