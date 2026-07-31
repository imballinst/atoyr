import type { SessionMode, SessionTopic } from '~/api/gen';

interface ModeBannerProps {
  mode: SessionMode;
  topic: SessionTopic;
}

const TOPIC_LABELS: Record<SessionTopic, string> = {
  'english-words': 'English Words',
  'indonesian-politician-quotes': 'Indonesian Politician Quotes',
};

export function ModeBanner({ mode, topic }: ModeBannerProps) {
  const isBlind = mode === 'blind';
  const displayMode = isBlind ? '🚫 BLIND MODE 🚫' : '🍦 Vanilla mode 🍦';

  return (
    <div
      className={`absolute top-0 w-full py-2 px-4 text-center text-sm font-bold ${
        isBlind ? 'bg-red-500/20 text-red-400' : 'bg-amber-500/20 text-amber-400'
      }`}
      role="status"
      aria-label={`${isBlind ? 'Blind' : 'Vanilla'} mode, ${TOPIC_LABELS[topic]} topic`}
    >
      {displayMode} &mdash; {TOPIC_LABELS[topic]}
    </div>
  );
}
