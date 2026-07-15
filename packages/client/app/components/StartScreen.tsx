import { HowToPlayModal } from './HowToPlayModal';
import { SettingsModal } from './SettingsModal';

interface StartScreenProps {
  onStart: () => void;
}

export function StartScreen({ onStart }: StartScreenProps) {
  return (
    <div className="rounded-2xl text-center w-full">
      <div className="w-full flex flex-col items-center mb-6">
        <h1 className="text-3xl text-left font-bold text-dark-interactive-primary">Atoyr</h1>

        <p className="text-dark-interactive-primary">
          <Title />
        </p>
      </div>

      <p className="text-sm text-dark-text-secondary mb-6 text-center">Unscramble 5-letter words as fast as you can!</p>

      <button
        type="button"
        onClick={() => onStart()}
        className="w-full py-3 px-6 text-base font-semibold bg-dark-interactive-primary text-white rounded-lg transition duration-200 hover:shadow-lg active:translate-y-0 shadow hover:bg-dark-interactive-hover"
        data-ga-label="ga-start-game-button"
      >
        Start Game
      </button>

      <div className="grid grid-cols-2 gap-2 mt-3 w-full">
        <HowToPlayModal />
        <SettingsModal />
      </div>
    </div>
  );
}

function Title() {
  const aTest = 'A test of your reflexes'.split(' ').map((word) => <TitleSegment key={word} word={word} />);
  const ofYourReflexes = ''.split(' ').map((word) => <TitleSegment key={word} word={word} />);

  return (
    <span className="flex flex-col gap-x-2 font-semibold">
      <span className="flex gap-x-1">{aTest}</span>
      <span className="flex gap-x-1">{ofYourReflexes}</span>
    </span>
  );
}

function TitleSegment({ word }: { word: string }) {
  const firstChar = <span>{word.charAt(0)}</span>;
  if (word.length === 1) return firstChar;

  return (
    <span>
      {firstChar}
      <span className="text-dark-text-secondary">{word.slice(1)}</span>
    </span>
  );
}
