# September 2026

## Week 3

- New feature: Switched the English words source from a hand-curated wordlist with LLM-generated definitions to [WordNet 3.1](https://wordnet.princeton.edu/download); definitions are now real dictionary glosses. As part of this, the old "English words" leaderboard will be _immortalized_ as "English words, July 2026" topic. This topic is only available for selection in the leaderboard and won't be playable anymore.
- Added more quotes to Indonesian Politician Quotes.

# August 2026

## Week 5

- New feature: leaderboard filter by "All time" and "this month".
- Adjusted the results screen so "Game over!" and the last word/quote appear on the same line, including some styling updates.
- Added labels to the leaderboard filters.
- Added date/time in every row of the leaderboard entry.
- Added more quotes for Indonesian Politican Quotes topic.
- Added references for Indonnesian Politician Quotes topic.

## Week 1

- New feature: Added "Share settings" in the Settings modal; the share URL is shown in a copyable field with a Copy button; visiting the link applies those settings automatically (unless a game is in progress).
- New feature: Added topics; two topic categories to choose from: English words (classic 5-letter scramble) and Indonesian politician quotes (fill-in-the-blank quotes). The topic selector lives in the Settings modal alongside the mode selector.
- Bug fix: Text-to-speech now always hides the word definition during gameplay.
- Bug fix: Fixed layout issues on smaller screens.
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
