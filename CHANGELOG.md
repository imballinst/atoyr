# July 2026

## Week 4

- New feature: Added Blind mode; a harder game mode that hides the word definition. A banner showing (Vanilla/Blind) below the navbar will show to indicate the game mode easily.
- Fixed a timer race condition where a wrong answer at exactly 1 second remaining could cause the game session to never be marked as finished, excluding your score from the leaderboard.

## Week 3

- New feature: Added changelog page, version indicator, and "New" badge in the navbar.
- New feature: Added How to Play and Settings modals accessible from the start screen.
- Adjusted modal triggers and border radius.

## Week 2

- New feature: Play again immediately after a game ends without returning home.
- New feature: Auto-submit now triggers when you fill all 5 letter slots.
- Adjusted Leaderboard percentile calculation to be more tolerant of edge cases.
- Fixed a race condition when finishing a game (score not saved).
