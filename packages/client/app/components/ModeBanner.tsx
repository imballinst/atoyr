import type { SessionMode } from '~/api/gen';

interface ModeBannerProps {
  mode: SessionMode;
}

export function ModeBanner({ mode }: ModeBannerProps) {
  const isBlind = mode === 'blind';

  return (
    <div
      className={`absolute top-0 w-full py-2 px-4 text-center text-sm font-bold ${
        isBlind ? 'bg-red-500/20 text-red-400' : 'bg-amber-500/20 text-amber-400'
      }`}
      role="status"
      aria-label={isBlind ? 'Blind mode' : 'Vanilla mode'}
    >
      {isBlind ? '🚫 BLIND MODE 🚫' : '🍦 Vanilla mode 🍦'}
    </div>
  );
}
