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

## Backend Implementation

### Technology Stack

The backend uses **Go 1.21** with the following frameworks:

- **Web Framework**: Gin (high-performance HTTP framework)
- **ORM**: GORM (Go Object-Relational Mapping for SQLite)
- **Database**: SQLite (serverless, self-contained)
- **Architecture**: Service-based with clear separation of concerns

### Why Go?

1. **Simplicity**: Explicit dependency injection, no decorator magic
2. **Performance**: Compiled binary with 5-40x better performance than Node.js alternatives
3. **Concurrency**: Native goroutines for timer management and concurrent requests
4. **Operational Efficiency**: Single executable deployment, minimal runtime requirements (~15MB memory vs 120MB+)
5. **Maintainability**: Straightforward code for CRUD + SSE use case

### Full Version Features

The full backend implementation includes:

1. **Session Management**: Server generates UUID for each game session
2. **SSE (Server-Sent Events)**: Real-time delivery of words and timing events to client
3. **Token Authentication**: Unique token per word to validate answer submissions
4. **HTTP API**: RESTful endpoints for game control and answer validation
5. **Persistent Storage**: SQLite database for game results and leaderboard rankings

See [02-backend.md](02-backend.md) for complete technical specification and implementation details.

## MVP Implementation

**Scope**: Client-side only with in-memory storage

**Status**: ✅ Complete and fully implemented

### Implemented Features

1. **Word Bucket Generation**
   - ✅ Filter dwyl word list for 5-letter words
   - ✅ Generate definitions via LLM script
   - ✅ Output to `packages/client/src/data/words.json`

2. **Game Logic**
   - ✅ Timer management with penalty tracking (30 seconds base)
   - ✅ Auto voice mode: +5 seconds per word when enabled
   - ✅ Word scrambling with edit distance calculation
   - ✅ Answer validation and scoring
   - ✅ No-skip retry mechanism with unlimited attempts
   - ✅ Leaderboard cookie persistence (7-day expiry)
   - ✅ Session word deduplication (no repeats per game)

3. **UI Components**
   - ✅ StartScreen: Game initialization with auto-voice toggle
   - ✅ GameScreen: Active gameplay with timer, scrambled word, definition, keyboard input
   - ✅ ResultsScreen: Final score, accuracy, leaderboard display
   - ✅ Wordle-style on-screen keyboard (full QWERTY layout)
   - ✅ Speaker button for manual text-to-speech
   - ✅ Responsive design (mobile-first, max 430px)

4. **State Management**
   - ✅ React hooks for game state, timer, leaderboard
   - ✅ Cookie-based leaderboard persistence (7-day expiry)
   - ✅ localStorage for auto-voice preference
   - ✅ Set-based tracking for session word usage

5. **Accessibility Features**
   - ✅ SpeechSynthesis: Automatically reads each letter when auto-voice enabled
   - ✅ Speaker button: Manual trigger to re-read letters
   - ✅ Physical keyboard support alongside on-screen keyboard
   - ✅ Screen reader friendly markup
   - ✅ Dark theme with high contrast colors

6. **Testing**
   - ✅ Unit tests for scrambling algorithm
   - ✅ Unit tests for edit distance calculation
   - ✅ Unit tests for answer validation
   - ✅ Vitest configuration for client and shared packages

## Future Enhancements

- Multi-player support
- Difficulty levels
- Custom word lists
- Analytics and statistics tracking
- Cross-session word tracking (to avoid repeats across plays)
- Redis caching for leaderboard optimization
- gRPC for high-performance internal APIs
- Serverless deployment options (AWS Lambda, Cloud Functions)
- Microservices architecture separation
