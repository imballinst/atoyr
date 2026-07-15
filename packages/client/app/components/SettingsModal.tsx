import { useState } from 'react';

import { readStoredAutoVoice, writeStoredAutoVoice } from '~/lib/auto-voice';

import { SharedDialog } from './Dialog';

export function SettingsModal() {
  const [autoVoice, setAutoVoice] = useState(readStoredAutoVoice);

  const handleAutoVoiceChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.checked;
    setAutoVoice(newValue);
    writeStoredAutoVoice(newValue);
  };

  return (
    <SharedDialog
      title="Settings"
      trigger={
        <button
          type="button"
          className="w-full py-2 px-4 text-sm font-medium text-dark-text-secondary border border-dark-border-primary rounded transition duration-200 hover:bg-dark-bg-tertiary hover:text-dark-text-primary"
          data-ga-label="ga-settings-button"
        >
          Settings
        </button>
      }
    >
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <input type="checkbox" id="auto-voice" checked={autoVoice} onChange={handleAutoVoiceChange} className="w-4 h-4 cursor-pointer" />
          <label htmlFor="auto-voice" className="text-sm text-left text-dark-text-secondary cursor-pointer flex-1">
            Enable automatic text-to-speech
          </label>
        </div>

        <p className="text-xs text-dark-text-secondary italic text-left">
          Uses your browser's built-in text-to-speech functionality. You will get 5 extra seconds for each word, but the scrambled letters
          will only be shown for screen readers.
        </p>
      </div>
    </SharedDialog>
  );
}
