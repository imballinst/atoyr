const STORAGE_KEY = 'atoyr_auto_voice:v1';
const LEGACY_STORAGE_KEY = 'atoyr_auto_voice';

function parseAutoVoice(value: string | null): boolean {
  if (value === null) return false;

  try {
    const parsed = JSON.parse(value);
    return typeof parsed === 'boolean' ? parsed : false;
  } catch {
    return false;
  }
}

export function readStoredAutoVoice(): boolean {
  if (typeof window === 'undefined') return false;

  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored !== null) {
    return parseAutoVoice(stored);
  }

  const legacy = localStorage.getItem(LEGACY_STORAGE_KEY);
  const value = parseAutoVoice(legacy);

  localStorage.setItem(STORAGE_KEY, JSON.stringify(value));
  localStorage.removeItem(LEGACY_STORAGE_KEY);

  return value;
}

export function writeStoredAutoVoice(value: boolean): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(value));
}
