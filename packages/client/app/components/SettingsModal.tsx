import { useId, type ReactNode } from 'react';

import type { SessionMode, SessionTopic } from '~/api/gen';
import { type LatestSchema } from '~/lib/settings';

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
          <p className="text-xs text-dark-text-secondary text-left">Some modes are not available with the current topic.</p>
        )}

        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between gap-x-2">
            <div className="text-sm">Mode</div>

            <div>
              <label className="text-sm">Mode</label>

              {isIndonesianTopic ? (
                <span className="text-sm ml-4 text-dark-text-secondary">Vanilla</span>
              ) : (
                <select
                  id={modeId}
                  className="text-sm"
                  onChange={(e) => {
                    updateSettings({ mode: e.target.value as SessionMode });
                  }}
                  value={settings.mode}
                >
                  <option value="vanilla">Vanilla</option>
                  <option value="blind">Blind</option>
                </select>
              )}
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
            <div className="text-sm">Topic</div>

            <div>
              <label htmlFor={topicId} className="sr-only">
                Topic
              </label>

              <select id={topicId} className="text-sm" onChange={handleTopicChange} value={settings.topic}>
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
      </div>
    </SharedDialog>
  );
}
