import z from 'zod';

export const LATEST_STORAGE_KEY = 'atoyr:settings:v1';

const V1Schema = z.boolean();
const V1SettingsSchema = z.object({
  autoVoice: z.boolean(),
  mode: z.enum(['vanilla', 'blind']),
});
type V1SettingsSchema = z.infer<typeof V1SettingsSchema>;

export type LatestSchema = V1SettingsSchema;

export const DEFAULT_LATEST_SCHEMA: V1SettingsSchema = {
  autoVoice: false,
  mode: 'vanilla',
};

const MIGRATIONS = {
  atoyr_auto_voice: {
    schema: z.boolean(),
    up: (cur: boolean) => cur satisfies z.infer<typeof V1Schema>,
  },
  'atoyr_auto_voice:v1': {
    schema: z.boolean(),
    up: (cur: boolean) => ({ autoVoice: cur, mode: 'vanilla' }) satisfies V1SettingsSchema,
  },
  'atoyr:settings:v1': {
    schema: V1SettingsSchema,
    up: null,
  },
};

function parseSettings(value: string | null): V1SettingsSchema {
  if (value === null) return DEFAULT_LATEST_SCHEMA;

  try {
    const parsed = JSON.parse(value);
    return V1SettingsSchema.parse(parsed);
  } catch {
    return DEFAULT_LATEST_SCHEMA;
  }
}

export function readStoredSettings(): z.infer<typeof V1SettingsSchema> {
  try {
    const latestStored = localStorage.getItem(LATEST_STORAGE_KEY);
    if (latestStored !== null) {
      return parseSettings(latestStored);
    }

    // If no latest stored, then find out old values.
    const entries = Object.entries(MIGRATIONS);
    const finalMigratedData = entries.reduce((migrated, [curKey, curValue]) => {
      if (curValue.up === null) return migrated;
      if (migrated !== null) return curValue.up(migrated);

      const stored = localStorage.getItem(curKey);
      if (stored === null) return null;

      const parsed = curValue.schema.safeParse(JSON.parse(stored));
      if (parsed.success) return parsed.data;

      return null;
    }, null as any);

    if (finalMigratedData === null) return DEFAULT_LATEST_SCHEMA;

    writeStoredSettings(finalMigratedData);
    return finalMigratedData;
  } catch (err) {
    console.error(err);

    return DEFAULT_LATEST_SCHEMA;
  } finally {
    // In any case, during errors, cleanup old keys.
    for (const key in MIGRATIONS) {
      if (key !== LATEST_STORAGE_KEY) localStorage.removeItem(key);
    }
  }
}

export function writeStoredSettings(value: V1SettingsSchema): void {
  localStorage.setItem(LATEST_STORAGE_KEY, JSON.stringify(value));
}
