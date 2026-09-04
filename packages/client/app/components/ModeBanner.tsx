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
  const emoji = isBlind ? '🚫' : '🍦';
  const modeLabel = isBlind ? 'BLIND MODE' : 'Vanilla mode';
  const modeLabelAndTopic = `${modeLabel} + ${TOPIC_LABELS[topic]}`;

  return (
    <div
      className={`absolute top-0 w-full py-1 text-center font-bold ${
        isBlind ? 'bg-red-500/20 text-red-400' : 'bg-amber-500/20 text-amber-400'
      } ${modeLabelAndTopic.length > 40 ? 'text-xs' : 'text-sm'}`}
      role="status"
      aria-label={`${isBlind ? 'Blind' : 'Vanilla'} mode, ${TOPIC_LABELS[topic]} topic`}
    >
      {emoji} {modeLabelAndTopic} {emoji}
    </div>
  );
}
