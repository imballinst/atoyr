# July 2026

## Week 4

- New feature: Added Blind mode — a harder game mode that hides the word definition.
- Added mode banner (Vanilla/Blind) below the navbar on the home page and game screen.
- Leaderboard percentile comparisons now respect the game mode.
- Fix: Definition is now omitted from all server responses in Blind mode (not just the client-side render).
- Fix: "Speak letters" no longer reads the definition in Blind mode.
- Fix: Settings auto-voice toggle now uses a proper accessible label.

## Week 3

- New feature: Added changelog page, version indicator, and "New" badge in the navbar.
- New feature: Added How to Play and Settings modals accessible from the start screen.
- Adjusted modal triggers and border radius.

## Week 2

- New feature: Play again immediately after a game ends without returning home.
- New feature: Auto-submit now triggers when you fill all 5 letter slots.
- Adjusted Leaderboard percentile calculation to be more tolerant of edge cases.
- Fixed a race condition when finishing a game (score not saved).
