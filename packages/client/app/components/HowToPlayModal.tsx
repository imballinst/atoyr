import { SharedDialog } from './Dialog';

export function HowToPlayModal() {
  return (
    <SharedDialog
      title="How to play"
      trigger={
        <button
          type="button"
          className="py-2 px-4 text-sm font-medium text-dark-text-secondary border border-dark-border-primary rounded transition duration-200 hover:bg-dark-bg-tertiary hover:text-dark-text-primary"
          data-ga-label="ga-how-to-play-button"
        >
          How to play
        </button>
      }
    >
      <ul className="text-sm text-dark-text-secondary space-y-3">
        {['You have 30 seconds', 'Wrong answers cost 1 second', 'Type using external keyboard or on-screen buttons'].map((rule, idx) => (
          <li key={idx} className="flex gap-2">
            <span className="text-dark-interactive-success font-bold shrink-0">✓</span>
            <span>{rule}</span>
          </li>
        ))}
      </ul>
    </SharedDialog>
  );
}
