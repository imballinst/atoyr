# August 2026

## Week 1

- New feature: Added "Share settings" in the Settings modal; the share URL is shown in a copyable field with a Copy button; visiting the link applies those settings automatically (unless a game is in progress).
- New feature: Added topics; two topic categories to choose from — English words (classic 5-letter scramble) and Indonesian politician quotes (fill-in-the-blank quotes). The topic selector lives in the Settings modal alongside the mode selector. Quote-topic entries get a "Quotes" badge on the leaderboard; leaderboards are partitioned per-topic so scores from different topics are never compared.
- Bug fix: Auto voice now hides the word definition during gameplay so the timer bonus is correctly applied when the voice reads the definition aloud.
- Bug fix: Fixed layout issues on smaller screens; the results screen header is now more compact and the in-game and results stats share a single stats bar.
- Adjusted the results screen stat cards (score, accuracy, best streak) into a single-row panel.

# July 2026

## Week 3

- New feature: Added Blind mode; a harder game mode that hides the word definition. A banner showing (Vanilla/Blind) below the navbar will show to indicate the game mode easily.
- New feature: Added changelog page, version indicator, and "New" badge in the navbar.
- New feature: Added How to Play and Settings modals accessible from the start screen.
- Bug fix: Fixed a timer race condition where a wrong answer at exactly 1 second remaining could cause the game session to never be marked as finished, excluding your score from the leaderboard.
- Adjusted modal triggers and border radius.

## Week 2

- New feature: Play again immediately after a game ends without returning home.
- New feature: Auto-submit now triggers when you fill all 5 letter slots.
- Bug fix: Fixed a race condition when finishing a game (score not saved).
- Adjusted Leaderboard percentile calculation to be more tolerant of edge cases.
