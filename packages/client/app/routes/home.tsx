import { Loader2Icon } from 'lucide-react';
import { useLoaderData } from 'react-router';

import { useGame } from '~/api/hooks';
import { GameScreen } from '~/components/GameScreen';
import { ResultsScreen } from '~/components/ResultsScreen';
import { StartScreen } from '~/components/StartScreen';
import { hasGameEnded } from '~/lib/game';
import { readStoredSettings } from '~/lib/settings';

export function meta() {
  return [{ title: 'Game | Atoyr' }, { name: 'description', content: 'Test your reflexes, climb the leaderboard.' }];
}

export function clientLoader() {
  return { shouldFetch: !hasGameEnded(), settings: readStoredSettings() };
}

export default function Home() {
  const { shouldFetch, settings } = useLoaderData<typeof clientLoader>();
  const { state, startGame, submitAnswer, playAgain, resetGame, updateSettings } = useGame(shouldFetch, settings);

  if (state.phase === 'resuming') {
    return (
      <div className="w-full h-full text-dark-text-primary flex flex-col items-center justify-center gap-y-2">
        <Loader2Icon className="animate-spin" />
        <div>Resuming your game...</div>
      </div>
    );
  }

  if (state.phase === 'idle') {
    return <StartScreen onStart={() => startGame(state.settings)} settings={state.settings} updateSettings={updateSettings} />;
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
        onPlayAgain={playAgain}
        onBackToHome={resetGame}
        correctAttemptTimestamps={state.correctAttemptTimestamps}
        settings={state.settings}
        onUpdateSettings={updateSettings}
      />
    );
  }

  return null;
}
