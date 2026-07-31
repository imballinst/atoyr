import z from 'zod';

export const LATEST_STORAGE_KEY = 'atoyr:settings:v2';

const V1Schema = z.boolean();
const V1SettingsSchema = z.object({
  autoVoice: z.boolean(),
  mode: z.enum(['vanilla', 'blind']),
});
export type V1SettingsSchema = z.infer<typeof V1SettingsSchema>;
const V2SettingsSchema = z.object({
  autoVoice: z.boolean(),
  mode: z.enum(['vanilla', 'blind']),
  topic: z.enum(['english-words', 'indonesian-politician-quotes']),
});
type V2SettingsSchema = z.infer<typeof V2SettingsSchema>;

export type LatestSchema = V2SettingsSchema;

export const DEFAULT_LATEST_SCHEMA: V2SettingsSchema = {
  autoVoice: false,
  mode: 'vanilla',
  topic: 'english-words',
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
    up: (cur: V1SettingsSchema) => ({ ...cur, topic: 'english-words' }) satisfies V2SettingsSchema,
  },
};

function parseSettings(value: string | null): V2SettingsSchema {
  if (value === null) return DEFAULT_LATEST_SCHEMA;

  try {
    const parsed = JSON.parse(value);
    return V2SettingsSchema.parse(parsed);
  } catch {
    return DEFAULT_LATEST_SCHEMA;
  }
}

export function readStoredSettings(): z.infer<typeof V2SettingsSchema> {
  try {
    const latestStored = localStorage.getItem(LATEST_STORAGE_KEY);
    if (latestStored !== null) {
      return parseSettings(latestStored);
    }

    const entries = Object.entries(MIGRATIONS);
    const finalMigratedData = entries.reduce((migrated, [curKey, curValue]) => {
      if (curValue.up === null) return migrated;
      if (migrated !== null) return curValue.up(migrated);

      const stored = localStorage.getItem(curKey);
      if (stored === null) return null;

      const parsed = curValue.schema.safeParse(JSON.parse(stored));
      if (parsed.success) return (curValue.up as any)(parsed.data);

      return null;
    }, null as any);

    if (finalMigratedData === null) return DEFAULT_LATEST_SCHEMA;

    writeStoredSettings(finalMigratedData);
    return finalMigratedData;
  } catch (err) {
    console.error(err);

    return DEFAULT_LATEST_SCHEMA;
  } finally {
    for (const key in MIGRATIONS) {
      if (key !== LATEST_STORAGE_KEY) localStorage.removeItem(key);
    }
  }
}

export function writeStoredSettings(value: LatestSchema): void {
  localStorage.setItem(LATEST_STORAGE_KEY, JSON.stringify(value));
}
