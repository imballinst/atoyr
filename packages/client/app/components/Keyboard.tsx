export function Keyboard({
  answer,
  onBackspace,
  onClick,
  onSubmit,
}: {
  answer: string;
  onClick: (letter: string) => void;
  onBackspace: () => void;
  onSubmit: () => void;
}) {
  return (
    <div className="space-y-2">
      <div className="w-full flex flex-col gap-2">
        <div className="grid grid-cols-10 gap-1 w-full">
          {['Q', 'W', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P'].map((key) => (
            <button
              key={key}
              className="aspect-square bg-dark-bg-tertiary border-2 border-dark-border-primary rounded font-semibold text-xs transition duration-200 text-dark-text-primary hover:bg-dark-interactive-primary hover:text-white hover:-translate-y-0.5 hover:shadow-lg active:translate-y-0 disabled:opacity-50 disabled:cursor-not-allowed"
              onClick={() => onClick(key)}
              disabled={answer.length >= 5}
            >
              {key}
            </button>
          ))}
        </div>
        <div className="grid grid-cols-9 gap-1 w-full">
          {['A', 'S', 'D', 'F', 'G', 'H', 'J', 'K', 'L'].map((key) => (
            <button
              key={key}
              className="aspect-square bg-dark-bg-tertiary border-2 border-dark-border-primary rounded font-semibold text-xs transition duration-200 text-dark-text-primary hover:bg-dark-interactive-primary hover:text-white hover:-translate-y-0.5 hover:shadow-lg active:translate-y-0 disabled:opacity-50 disabled:cursor-not-allowed"
              onClick={() => onClick(key)}
              disabled={answer.length >= 5}
            >
              {key}
            </button>
          ))}
        </div>
        <div className="grid grid-cols-8 gap-1 w-full">
          {['Z', 'X', 'C', 'V', 'B', 'N', 'M'].map((key) => (
            <button
              key={key}
              className="aspect-square bg-dark-bg-tertiary border-2 border-dark-border-primary rounded font-semibold text-xs transition duration-200 text-dark-text-primary hover:bg-dark-interactive-primary hover:text-white hover:-translate-y-0.5 hover:shadow-lg active:translate-y-0 disabled:opacity-50 disabled:cursor-not-allowed"
              onClick={() => onClick(key)}
              disabled={answer.length >= 5}
            >
              {key}
            </button>
          ))}

          <button
            className="aspect-square bg-dark-bg-primary border-2 border-dark-border-primary rounded font-semibold text-xs transition duration-200 text-dark-text-primary hover:bg-dark-interactive-primary hover:text-white hover:-translate-y-0.5 hover:shadow-lg active:translate-y-0 disabled:opacity-50 disabled:cursor-not-allowed"
            onClick={onBackspace}
            disabled={answer.length >= 5}
            aria-label="Backspace"
          >
            ←
          </button>
        </div>
      </div>

      <button
        onClick={onSubmit}
        disabled={answer.length === 0}
        className="w-full py-3 bg-dark-interactive-success text-white font-semibold text-sm rounded transition duration-200 hover:bg-green-600 hover:-translate-y-0.5 hover:shadow-lg active:translate-y-0 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        Submit
      </button>
    </div>
  );
}
