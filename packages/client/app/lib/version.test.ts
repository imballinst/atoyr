import { describe, expect, it } from 'vitest';

import { computeCombinedVersion, getClientVersion, getServerVersion, shouldShowBadge } from './version';

describe('computeCombinedVersion', () => {
  it('returns Dev when version is undefined', () => {
    expect(computeCombinedVersion(undefined)).toBe('Dev');
  });

  it('adds semver components component-wise', () => {
    expect(computeCombinedVersion('0.1.0-0.2.1')).toBe('v0.3.1');
    expect(computeCombinedVersion('1.2.3-4.5.6')).toBe('v5.7.9');
  });

  it('does not carry over between components', () => {
    expect(computeCombinedVersion('9.9.9-9.9.9')).toBe('v18.18.18');
  });

  it('returns the raw string when the format is invalid', () => {
    expect(computeCombinedVersion('1.0.0')).toBe('1.0.0');
    expect(computeCombinedVersion('1.0-2.0')).toBe('1.0-2.0');
    expect(computeCombinedVersion('a.b.c-x.y.z')).toBe('a.b.c-x.y.z');
  });
});

describe('getClientVersion', () => {
  it('returns Dev when version is undefined', () => {
    expect(getClientVersion(undefined)).toBe('Dev');
  });

  it('returns the client part of the version', () => {
    expect(getClientVersion('0.1.0-0.2.1')).toBe('0.1.0');
  });

  it('returns the raw string when the format is invalid', () => {
    expect(getClientVersion('invalid')).toBe('invalid');
  });
});

describe('getServerVersion', () => {
  it('returns Dev when version is undefined', () => {
    expect(getServerVersion(undefined)).toBe('Dev');
  });

  it('returns the server part of the version', () => {
    expect(getServerVersion('0.1.0-0.2.1')).toBe('0.2.1');
  });

  it('returns the last segment when the format is invalid', () => {
    expect(getServerVersion('invalid')).toBe('invalid');
    expect(getServerVersion('a-b-c')).toBe('c');
  });
});

describe('shouldShowBadge', () => {
  const currentVersion = 'v0.3.1';
  const releaseDate = new Date('2026-07-01');
  const beforeOneWeek = new Date('2026-07-07');
  const afterOneWeek = new Date('2026-07-08');

  it('shows the badge when the last seen version differs', () => {
    expect(shouldShowBadge(currentVersion, 'v0.2.0', '2026-07-01T00:00:00.000Z', afterOneWeek, releaseDate)).toBe(true);
  });

  it('hides the badge when it has already been dismissed for the current version', () => {
    expect(shouldShowBadge(currentVersion, currentVersion, '2026-07-08T00:00:00.000Z', afterOneWeek, releaseDate)).toBe(false);
  });

  it('hides the badge before the one-week reminder window', () => {
    expect(shouldShowBadge(currentVersion, currentVersion, null, beforeOneWeek, releaseDate)).toBe(false);
  });

  it('shows the badge after the one-week reminder window if not dismissed', () => {
    expect(shouldShowBadge(currentVersion, currentVersion, null, afterOneWeek, releaseDate)).toBe(true);
  });

  it('hides the badge when release date is unknown and the version has been seen', () => {
    expect(shouldShowBadge(currentVersion, currentVersion, null, afterOneWeek, null)).toBe(false);
  });
});
