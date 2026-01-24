# Core

## Game Overview

"A Test of Your Reflexes" (Atoyr) is a word unscrambling game where players are shown scrambled 5-letter words with definitions and must unscramble them within a time limit.

## Features

### Core Gameplay

1. **Timer**: Fixed 30-second countdown that runs continuously throughout the game
   - Wrong answers incur 1-second penalty
   - Timer stops when it reaches 0

2. **Word Bucket**: Collection of 5-letter words with definitions
   - Source: dwyl/english-words (https://github.com/dwyl/english-words/blob/master/words.txt)
   - Definitions: LLM-generated (paraphrased, non-formal, max 5 words each)
   - Format: `{ word: string; definition: string }`
   - Minimum 100-500 words for MVP

3. **Word Scrambling**: Maximize difficulty through strategic shuffling
   - Generate 10 random shuffles of the word
   - Calculate edit distance (Levenshtein) for each
   - Select the shuffle with **highest** edit distance (hardest to unscramble)
   - If all shuffles match the original, skip the word and select another
   - No word repeats within a single session

4. **Answer Validation**
   - Compare player input (case-insensitive, trimmed) to original word
   - Correct answer: increment score, load next word
   - Incorrect answer: 1-second time penalty, show feedback, allow unlimited retries (no skip)

5. **Scoring & Results**
   - Track both correct count and total attempts
   - Display accuracy (correct/attempts)
   - Leaderboard: stored in cookie as JSON, expires after 1 week
   - Save result on game end with timestamp

### Accessibility Features

- **SpeechSynthesis**: Automatically reads each letter when a new word appears
- **Speaker Button**: Manual trigger to re-read letters
- **Input Methods**: Both physical keyboard and on-screen Wordle-style keyboard
- **Mobile-First Design**:
  - Max container width: 500px
  - Keyboard optimized for 300px minimum, scales up to 430px maximum
  - Desktop browsers display at mobile viewport size

## Backend Implementation (Full Version)

_Note: MVP is client-only. Full version will include:_

1. Session Management: Server generates UUID for each game session
2. SSE (Server-Sent Events): Delivers words and timing events to client
3. Token Authentication: Unique token per word to validate requests
4. HTTP API: Endpoint to submit answers and receive validation
5. Database: SQLite storage for game results and leaderboard

## MVP Implementation

**Scope**: Client-side only with in-memory storage

1. Word Bucket Generation
   - Filter dwyl word list for 5-letter words
   - Generate definitions via LLM script
   - Output to `packages/client/src/data/words.json`

2. Game Logic
   - Timer management with penalty tracking
   - Word scrambling with edit distance calculation
   - Answer validation and scoring
   - No-skip retry mechanism

3. UI Components
   - StartScreen: Game initialization
   - GameScreen: Active gameplay with timer, word, keyboard input
   - ResultsScreen: Final score, accuracy, leaderboard
   - Wordle-style on-screen keyboard
   - SpeechSynthesis integration

4. State Management
   - React hooks for game state, timer, leaderboard
   - Cookie-based leaderboard persistence (1 week)

5. Testing
   - Unit tests for scrambling, edit distance, validation logic
   - State management tests
   - Timer behavior tests

## Future Enhancements

- Backend server with SSE integration
- Persistent database (SQLite)
- Multi-player support
- Difficulty levels
- Custom word lists
- Analytics and statistics tracking
- Cross-session word tracking (to avoid repeats across plays)
