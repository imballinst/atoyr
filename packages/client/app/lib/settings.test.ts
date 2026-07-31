import { beforeEach, describe, expect, it } from 'vitest';

import { LATEST_STORAGE_KEY, readStoredSettings, type V1SettingsSchema } from './settings';

describe('settings', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('migrates v1 settings to v2 with topic: english-words', () => {
    localStorage.setItem('atoyr:settings:v1', JSON.stringify({ autoVoice: true, mode: 'blind' } satisfies V1SettingsSchema));

    const settings = readStoredSettings();

    expect(settings).toEqual({ autoVoice: true, mode: 'blind', topic: 'english-words' });
    expect(localStorage.getItem(LATEST_STORAGE_KEY)).toBe(JSON.stringify(settings));
  });

  it('returns defaults when no stored settings exist', () => {
    const settings = readStoredSettings();

    expect(settings).toEqual({ autoVoice: false, mode: 'vanilla', topic: 'english-words' });
  });

  it('reads existing v2 settings directly', () => {
    localStorage.setItem(LATEST_STORAGE_KEY, JSON.stringify({ autoVoice: true, mode: 'blind', topic: 'indonesian-politician-quotes' }));

    const settings = readStoredSettings();

    expect(settings).toEqual({ autoVoice: true, mode: 'blind', topic: 'indonesian-politician-quotes' });
  });
});
