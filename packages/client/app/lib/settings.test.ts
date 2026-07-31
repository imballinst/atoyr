import { beforeEach, describe, expect, it } from 'vitest';

import {
  decodeSettings,
  encodeSettings,
  LATEST_STORAGE_KEY,
  readStoredSettings,
  type LatestSchema,
  type V1SettingsSchema,
} from './settings';

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

describe('encodeSettings / decodeSettings', () => {
  const settings: LatestSchema = { autoVoice: true, mode: 'blind', topic: 'indonesian-politician-quotes' };

  it('round-trips a full settings object', () => {
    const encoded = encodeSettings(settings);
    const decoded = decodeSettings(encoded);

    expect(decoded).toEqual(settings);
  });

  it('returns null for non-base64 input', () => {
    expect(decodeSettings('!!!not-base64')).toBeNull();
  });

  it('returns null for base64 of invalid JSON', () => {
    expect(decodeSettings(btoa('nope'))).toBeNull();
  });

  it('returns null when a required field is missing', () => {
    const partial = btoa(JSON.stringify({ autoVoice: false, mode: 'vanilla' }));

    expect(decodeSettings(partial)).toBeNull();
  });

  it('returns null for an unknown mode value', () => {
    const bad = btoa(JSON.stringify({ autoVoice: false, mode: 'nonexistent', topic: 'english-words' }));

    expect(decodeSettings(bad)).toBeNull();
  });

  it.each([
    { field: 'autoVoice', label: 'false', value: { autoVoice: false, mode: 'vanilla' as const, topic: 'english-words' as const } },
    { field: 'mode', label: 'vanilla', value: { autoVoice: false, mode: 'vanilla' as const, topic: 'english-words' as const } },
    { field: 'topic', label: 'english-words', value: { autoVoice: false, mode: 'vanilla' as const, topic: 'english-words' as const } },
  ])('encodes settings containing $field: $label without error', ({ value }) => {
    const encoded = encodeSettings(value);
    const decoded = decodeSettings(encoded);

    expect(decoded).toEqual(value);
  });
});
