# Changelog Page & Version Indicator

## Context

The navbar currently shows a `?` icon that opens a Popover displaying the raw client and server versions. We want to turn the version into a more visible UI element, add a dedicated changelog page, and nudge players to check it when something new ships.

---

## High-Level Goals

1. **Show a single combined version** in the navbar instead of the `?` icon.
2. **Reveal the breakdown on hover** — client version, server version, and a link to the changelog.
3. **Add a "New" badge** that appears when the player likely has unseen changes, and disappears once they open the version panel.
4. **Add a `/changelog` page** written in Markdown and organized by month/week, not by version number.

---

## High-Level Design

### 1. Combined Version Indicator

The navbar version chip shows one derived version string built from the client and server versions. The rule is simple component-wise addition:

- Client `0.1.0` + Server `0.2.1` → displayed as `v0.3.1`.

The chip replaces the `?` icon in the top-right of the navbar. It is still a small, non-intrusive element, but it now communicates that the game has a version.

### 2. Hover Panel

Hovering the version chip opens a small panel that shows:

- `Client version: {client}`
- `Server version: {server}`
- A link to `/changelog` (e.g., "See what changed")

The panel closes when the pointer leaves. The panel content is the same regardless of whether the "New" badge is visible.

### 3. "New" Badge

A small "New" label is rendered on the top-right corner of the version chip.

It is shown when **either** of these is true:

1. The version stored in `localStorage` is different from the current combined version. This catches fresh releases and first-time visitors.
2. The stored version equals the current combined version **and** the current date is at least one week after the latest release date recorded in the changelog **and** the player has not yet opened the version panel during this release cycle.

Opening the hover panel dismisses the badge immediately and writes the current version and the current timestamp to `localStorage`. Dismissing the badge once per release is enough; it should not return until the next release or until the one-week reminder window reopens for a future release.

### 4. Changelog Page

A new `/changelog` route renders the changelog from a Markdown file. The changelog is organized like a diary, not a version list:

- Top-level heading (`#`) is a month and year, e.g., `July 2026`.
- Second-level heading (`##`) is a week within that month, e.g., `Week 1`, `Week 2`, `Week 3`, `Week 4`.
- Bullet points under each week describe the changes.

No version numbers appear in the changelog headings or body. The changelog should only contain user-facing, essential changes. Internal refactors, dependency bumps, and infrastructure-only work are omitted unless they visibly affect the player.

---

## Low-Level Design

### Data Sources

- `import.meta.env.VERSION` is already injected as `"{client}-{server}"` (e.g., `"0.1.0-0.2.1"`).
- Add a `CHANGELOG.md` file at the repository root.
- The changelog file has no YAML frontmatter. The latest release date is inferred from the first heading pair:
  - First `h1` is parsed as `Month Year` (e.g., `July 2026`).
  - First `h2` under that `h1` is parsed as `Week N` (e.g., `Week 2`).
  - Release date = first day of the month + `(N - 1) * 7` days.

  Example:

  ```markdown
  # July 2026

  ## Week 2

  - Added a changelog page.
  - Improved the version indicator in the navbar.
  ```

  The latest release date for the example above is `2026-07-08`.

- The latest release date used for the "New" badge logic is derived from the first `h1`/`h2` pair in `CHANGELOG.md`, not from frontmatter.

### Version Calculation

```text
combinedMajor = clientMajor + serverMajor
combinedMinor = clientMinor + serverMinor
combinedPatch = clientPatch + serverPatch

result = "v{combinedMajor}.{combinedMinor}.{combinedPatch}"
```

- Each component is parsed as an integer and added arithmetically.
- No carry-over logic. If the sum is `11`, display `11`.
- If parsing fails for any component, fall back to displaying the raw `VERSION` string.

### Local Storage Schema

Use these keys:

- `atoyr:version:lastSeen` — the last combined version string the player dismissed the badge for.
- `atoyr:version:dismissedAt` — ISO timestamp of the last dismissal.

### "New" Badge Logic

```text
const releaseDate = deriveReleaseDate(CHANGELOG)
const oneWeekAfterRelease = addDays(releaseDate, 7)

showBadge = (
  lastSeen !== currentVersion
) || (
  lastSeen === currentVersion &&
  dismissedAt === null &&
  now >= oneWeekAfterRelease
)
```

- `deriveReleaseDate` parses the first `h1` as `Month Year` and the first `h2` under it as `Week N`, then returns the month start plus `(N - 1) * 7` days.
- `dismissedAt === null` means the player has not opened the version panel while on the current version.
- If the changelog headings cannot be parsed, the second branch is skipped; the badge still shows when `lastSeen !== currentVersion`.
- If the badge is dismissed, write both `lastSeen` and `dismissedAt`.

### Interaction Flow

1. Player hovers over the version chip.
2. The hover panel opens and shows client/server versions and the changelog link.
3. The "New" badge is hidden.
4. `localStorage` is updated with `lastSeen = currentVersion` and `dismissedAt = now()`.
5. If the player clicks the changelog link, navigate to `/changelog`.

### Components

- `VersionChip` — renders the combined version string and the "New" badge.
- `VersionHoverCard` — uses Radix `HoverCard` to render the hover panel with client/server versions and the changelog link. The panel is hover-only; there is no click-to-open behavior.
- `NewBadge` — a small visual indicator.

### Route

Add to `packages/client/app/routes.ts`:

```ts
route('/changelog', 'routes/changelog.tsx'),
```

### Navbar Hover Effects

Apply a consistent hover treatment to all navbar links (`Play`, `Leaderboard`, `About`) in `PageLayout.tsx` so the version chip is not the only interactive element with hover feedback. Use a subtle Tailwind hover class such as `hover:text-dark-interactive-primary` or a light background change. The active/current page indicator (`font-bold`) must remain visible.

### Changelog Page Rendering

- `routes/changelog.tsx` reads the markdown content at build time (static import or Vite's `?raw` import).
- Render it using `marked`, following the existing project convention for static Markdown pages.
- The page should have a back link or rely on the existing navbar.

### Heading Validation

The changelog parser should validate that:

- The first heading is `h1` and matches the pattern `Month Year` (e.g., `July 2026`).
- The second-level headings are `h2` and match `Week N` where `N` is 1–4.
- Only user-facing changes are listed. This is a human convention, not enforced by code.

### AGENTS.md Update

After implementing this feature, update the project `AGENTS.md` to document the changelog convention:

- Changelog entries must be user-facing and essential.
- Use `Month Year` for `h1` and `Week N` for `h2`.
- Keep the newest month/week at the top of the file; the "New" badge derives the latest release date from the first `h1`/`h2` pair.
- Do not log internal refactors or dependency-only changes unless the player can see them.

---

## Out of Scope

- Automatic changelog generation from commits.
- Multiple release channels or nightly versions.
- Push notifications or toasts for new releases.
- Server-side rendering of the changelog (static build-time import is sufficient).

---

## Decisions

- **Hover-only version panel**: the version chip uses Radix `HoverCard`, not a click-open `Popover`. The panel opens on hover and closes when the pointer leaves.
- **No `releasedAt` frontmatter**: the latest release date is derived from the first `h1` (`Month Year`) and first `h2` (`Week N`) in `CHANGELOG.md`.
- **Navbar hover effects**: all navbar links receive the same hover feedback treatment as the version chip.
