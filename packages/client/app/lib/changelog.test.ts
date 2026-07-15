import { describe, expect, it } from 'vitest';

import { deriveReleaseDate } from './changelog';

describe('deriveReleaseDate', () => {
  it('derives the release date from the first h1 and h2 pair', () => {
    const markdown = `# July 2026

## Week 2

- Some change.
`;
    const result = deriveReleaseDate(markdown);
    expect(result).toEqual(new Date(2026, 6, 8));
  });

  it('returns null when there is no valid h1', () => {
    const markdown = `## Week 2

- Some change.
`;
    expect(deriveReleaseDate(markdown)).toBeNull();
  });

  it('returns null when there is no valid h2 after h1', () => {
    const markdown = `# July 2026

- Some change.
`;
    expect(deriveReleaseDate(markdown)).toBeNull();
  });

  it('returns null for an invalid month', () => {
    const markdown = `# Smarch 2026

## Week 2

- Some change.
`;
    expect(deriveReleaseDate(markdown)).toBeNull();
  });

  it('returns null for an invalid week', () => {
    const markdown = `# July 2026

## Week 0

- Some change.
`;
    expect(deriveReleaseDate(markdown)).toBeNull();
  });

  it('uses the first h1 and h2 only from descending-order changelog', () => {
    const markdown = `# August 2026

## Week 3

- Latest entry.

# July 2026

## Week 1

- Older entry.
`;
    expect(deriveReleaseDate(markdown)).toEqual(new Date(2026, 7, 15));
  });

  it('handles Week 5 correctly', () => {
    const markdown = `# July 2026

## Week 5

- Some change.
`;
    expect(deriveReleaseDate(markdown)).toEqual(new Date(2026, 6, 29));
  });

  it('clamps Week 5 when the month has fewer days', () => {
    const markdown = `# February 2026

## Week 5

- Some change.
`;
    expect(deriveReleaseDate(markdown)).toEqual(new Date(2026, 1, 28));
  });
});
