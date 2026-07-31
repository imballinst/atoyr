import { useEffect, useId, useRef, useState, type ReactNode } from 'react';

import type { SessionMode, SessionTopic } from '~/api/gen';
import { encodeSettings, type LatestSchema } from '~/lib/settings';

import { SharedDialog } from './Dialog';

interface Props {
  triggerText?: ReactNode;
  triggerClassnames?: string;
  settings: LatestSchema;
  updateSettings: (schema: Partial<LatestSchema>) => void;
}

export function SettingsModal({ triggerText = 'Settings', triggerClassnames = '', settings, updateSettings }: Props) {
  const modeId = useId();
  const topicId = useId();
  const isIndonesianTopic = settings.topic === 'indonesian-politician-quotes';
  const [copied, setCopied] = useState(false);
  const timeoutRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  const handleAutoVoiceChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.checked;
    updateSettings({ autoVoice: newValue });
  };

  const handleTopicChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const topic = e.target.value as SessionTopic;
    if (topic === 'indonesian-politician-quotes') {
      updateSettings({ topic, mode: 'vanilla' });
    } else {
      updateSettings({ topic });
    }
  };

  const handleShare = async () => {
    const url = `${location.origin}/?settings=${encodeSettings(settings)}`;

    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      clearTimeout(timeoutRef.current);
      timeoutRef.current = setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    return () => clearTimeout(timeoutRef.current);
  }, []);

  return (
    <SharedDialog
      title="Settings"
      trigger={
        <button
          type="button"
          aria-label="Settings"
          className={
            'h-full py-2 px-4 text-sm font-medium text-dark-text-secondary border border-dark-border-primary rounded transition duration-200 hover:bg-dark-bg-tertiary hover:text-dark-text-primary ' +
            triggerClassnames
          }
          data-ga-label="ga-settings-button"
        >
          {triggerText}
        </button>
      }
    >
      <div className="flex flex-col gap-y-6">
        {isIndonesianTopic && (
          <p className="text-xs text-dark-text-secondary text-left border rounded p-2 bg-amber-900 border-amber-900">
            Some modes are unavailable with the current topic.
          </p>
        )}

        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between gap-x-2">
            <div className="text-sm">Mode</div>

            <div>
              <select
                disabled={isIndonesianTopic}
                id={modeId}
                className={`text-sm${isIndonesianTopic ? ' cursor-not-allowed text-dark-text-muted' : ''}`}
                onChange={(e) => {
                  updateSettings({ mode: e.target.value as SessionMode });
                }}
                value={settings.mode}
              >
                <option value="vanilla">Vanilla</option>
                <option value="blind">Blind</option>
              </select>
            </div>
          </div>

          <p className="text-xs text-dark-text-secondary text-left">
            {settings.mode === 'vanilla'
              ? 'Default game mode. Each scramble word will have a definition as a clue.'
              : 'Harder game mode. Each scramble word will NOT have a definition as a clue.'}
          </p>
        </div>

        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between gap-x-2">
            <label htmlFor={topicId} className="text-sm">
              Topic
            </label>

            <div>
              <select
                id={topicId}
                className="text-sm text-right [text-align-last:right]"
                onChange={handleTopicChange}
                value={settings.topic}
              >
                <option value="english-words">English Words</option>
                <option value="indonesian-politician-quotes">Indonesian Politician Quotes</option>
              </select>
            </div>
          </div>

          <p className="text-xs text-dark-text-secondary text-left">
            {settings.topic === 'english-words'
              ? 'Rearrange scrambled letters into English words.'
              : 'Fill in the missing word in quotes attributed to Indonesian politicians. The words in this topic are purely Indonesian.'}
          </p>
        </div>

        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between gap-2">
            <div className="text-sm">Text-to-speech</div>

            <div className="flex items-center gap-x-2">
              <label htmlFor="auto-voice" className="text-sm text-left text-dark-text-secondary cursor-pointer flex-1">
                {settings.autoVoice ? 'Enabled' : 'Disabled'}
              </label>

              <input
                type="checkbox"
                id="auto-voice"
                checked={settings.autoVoice}
                onChange={handleAutoVoiceChange}
                className="w-4 h-4 cursor-pointer"
              />
            </div>
          </div>

          <p className="text-xs text-dark-text-secondary text-left">
            Uses your browser's built-in text-to-speech functionality. You will get 5 extra seconds for each word, but the scrambled letters
            will only be shown for screen readers.
          </p>
        </div>

        <div className="border-t border-dark-border-primary pt-4">
          <p className="text-xs text-dark-text-secondary text-left mb-2">Share your current settings:</p>

          <div className="flex">
            <input
              type="text"
              readOnly
              value={`${location.origin}/?settings=${encodeSettings(settings)}`}
              className="flex-1 text-sm bg-dark-bg-tertiary border border-r-0 border-dark-border-primary rounded-l px-3 py-2 text-dark-text-secondary outline-none"
              onFocus={(e) => e.target.select()}
            />
            <button
              type="button"
              onClick={handleShare}
              className="text-sm font-medium bg-dark-interactive-primary text-white rounded-r px-4 py-2 transition duration-200 hover:bg-dark-interactive-hover border border-dark-border-primary min-w-[85px] whitespace-nowrap"
            >
              {copied ? 'Copied!' : 'Copy'}
            </button>
          </div>
        </div>
      </div>
    </SharedDialog>
  );
}
