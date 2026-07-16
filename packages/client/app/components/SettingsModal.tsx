import { useId, type ReactNode } from 'react';

import type { SessionMode } from '~/api/gen';
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

  const handleAutoVoiceChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.checked;
    updateSettings({ autoVoice: newValue });
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
        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between gap-x-2">
            <div className="text-sm">Mode</div>

            <div>
              <label htmlFor={modeId} className="sr-only">
                Mode
              </label>

              <select
                className="text-sm"
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
