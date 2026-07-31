export function Keyboard({
  answer,
  wordLength,
  onBackspace,
  onClick,
}: {
  answer: string;
  wordLength: number;
  onClick: (letter: string) => void;
  onBackspace: () => void;
}) {
  return (
    <div className="space-y-2 w-full">
      <div className="w-full flex flex-col gap-2">
        <div className="grid grid-cols-10 gap-1 w-full">
          {['Q', 'W', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P'].map((key) => (
            <button
              key={key}
              type="button"
              className="aspect-square bg-dark-bg-tertiary border border-dark-border-primary rounded font-semibold text-xs transition duration-200 text-dark-text-primary hover:bg-dark-interactive-primary hover:text-white hover:shadow-lg active:translate-y-0 disabled:opacity-50 disabled:cursor-not-allowed"
              onPointerDown={(e) => {
                e.preventDefault();
                onClick(key);
              }}
              disabled={answer.length >= wordLength}
            >
              {key}
            </button>
          ))}
        </div>
        <div className="grid grid-cols-20 gap-1 w-full">
          <div className="col-span-1" aria-hidden />

          {['A', 'S', 'D', 'F', 'G', 'H', 'J', 'K', 'L'].map((key) => (
            <button
              key={key}
              type="button"
              className="col-span-2 aspect-square bg-dark-bg-tertiary border border-dark-border-primary rounded font-semibold text-xs transition duration-200 text-dark-text-primary hover:bg-dark-interactive-primary hover:text-white hover:shadow-lg active:translate-y-0 disabled:opacity-50 disabled:cursor-not-allowed"
              onClick={() => onClick(key)}
              disabled={answer.length >= wordLength}
            >
              {key}
            </button>
          ))}

          <div className="col-span-1" aria-hidden />
        </div>
        <div className="grid grid-cols-20 gap-1 w-full">
          <div className="col-span-3" aria-hidden />

          {['Z', 'X', 'C', 'V', 'B', 'N', 'M'].map((key) => (
            <button
              key={key}
              type="button"
              className="col-span-2 aspect-square bg-dark-bg-tertiary border border-dark-border-primary rounded font-semibold text-xs transition duration-200 text-dark-text-primary hover:bg-dark-interactive-primary hover:text-white hover:shadow-lg active:translate-y-0 disabled:opacity-50 disabled:cursor-not-allowed"
              onClick={() => onClick(key)}
              disabled={answer.length >= wordLength}
            >
              {key}
            </button>
          ))}

          <button
            type="button"
            className="col-span-3 bg-dark-bg-primary border border-dark-border-primary rounded font-semibold text-xs transition duration-200 text-dark-text-primary hover:bg-dark-interactive-primary hover:text-white hover:shadow-lg active:translate-y-0 disabled:opacity-50 disabled:cursor-not-allowed"
            onPointerDown={(e) => {
              e.preventDefault();
              onBackspace();
            }}
            aria-label="Backspace"
          >
            ←
          </button>
        </div>
      </div>
    </div>
  );
}
