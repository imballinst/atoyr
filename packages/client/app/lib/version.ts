const LAST_SEEN_KEY = 'atoyr:version:lastSeen';
const DISMISSED_AT_KEY = 'atoyr:version:dismissedAt';

export function computeCombinedVersion(version: string | undefined): string {
  if (!version) return 'Dev';

  const parts = version.split('-');
  if (parts.length !== 2) return version;

  const [clientPart, serverPart] = parts;
  const clientComponents = clientPart.split('.').map(Number);
  const serverComponents = serverPart.split('.').map(Number);

  if (clientComponents.length !== 3 || serverComponents.length !== 3) return version;
  if (clientComponents.some(isNaN) || serverComponents.some(isNaN)) return version;

  const major = clientComponents[0] + serverComponents[0];
  const minor = clientComponents[1] + serverComponents[1];
  const patch = clientComponents[2] + serverComponents[2];

  return `v${major}.${minor}.${patch}`;
}

function parseComponents(version: string): { client: number[]; server: number[] } | null {
  const parts = version.split('-');
  if (parts.length !== 2) return null;

  const client = parts[0].split('.').map(Number);
  const server = parts[1].split('.').map(Number);

  if (client.length !== 3 || server.length !== 3) return null;
  if (client.some(isNaN) || server.some(isNaN)) return null;

  return { client, server };
}

export function getClientVersion(version: string | undefined): string {
  if (!version) return 'Dev';
  const parsed = parseComponents(version);
  return parsed ? parsed.client.join('.') : version;
}

export function getServerVersion(version: string | undefined): string {
  if (!version) return 'Dev';
  const parsed = parseComponents(version);
  return parsed ? parsed.server.join('.') : (version.split('-').pop() ?? version);
}

export function readStoredVersionInfo() {
  if (typeof window === 'undefined') {
    return { lastSeen: null, dismissedAt: null };
  }

  return {
    lastSeen: localStorage.getItem(LAST_SEEN_KEY),
    dismissedAt: localStorage.getItem(DISMISSED_AT_KEY),
  };
}

export function writeStoredVersionInfo(lastSeen: string, dismissedAt: string): void {
  localStorage.setItem(LAST_SEEN_KEY, lastSeen);
  localStorage.setItem(DISMISSED_AT_KEY, dismissedAt);
}

export function shouldShowBadge(
  currentVersion: string,
  lastSeen: string | null,
  dismissedAt: string | null,
  now: Date,
  releaseDate: Date | null,
): boolean {
  if (lastSeen !== currentVersion) return true;

  if (lastSeen === currentVersion && dismissedAt === null && releaseDate !== null) {
    const oneWeekAfter = new Date(releaseDate);
    oneWeekAfter.setDate(oneWeekAfter.getDate() + 7);
    return now >= oneWeekAfter;
  }

  return false;
}
