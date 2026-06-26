# MVP Technical Specification

## Overview

The MVP implements a client-side only version of "A Test of Your Reflexes" (Atoyr). Players are shown scrambled 5-letter words with definitions and must unscramble them within a time limit.

## Architecture

### Client-Only Architecture

For MVP, all logic runs in the browser:

```
┌─────────────────────────────────────────────┐
│                   Client                    │
│  ┌─────────────┐  ┌──────────────────────┐  │
│  │  React UI   │  │   Game Logic         │  │
│  │             │◄─┤   - Timer            │  │
│  │  - Display  │  │   - Word Selection   │  │
│  │  - Input    │  │   - Scrambling       │  │
│  │  - Results  │  │   - Validation       │  │
│  └─────────────┘  └──────────────────────┘  │
│                           │                 │
│                   ┌───────▼───────┐         │
│                   │  Word Bucket  │         │
│                   │  (JSON/Static)│         │
│                   └───────────────┘         │
└─────────────────────────────────────────────┘
```

## Features

### 1. Word Bucket (Behind the Scenes)

**Goal**: Populate a bucket of 5-letter words with definitions.

**Implementation**:

- Use https://github.com/dwyl/english-words/blob/master/words.txt as the word source
- Generate definitions using an LLM script (paraphrased, non-formal wording, max 5 words)
- Store results in a static JSON file: `packages/client/src/data/words.json`
- Format:

  ```typescript
  interface WordEntry {
    word: string; // 5-letter word, lowercase
    definition: string; // Short definition, max 5 words, non-formal
  }

  type WordBucket = WordEntry[];
  ```

**Word Generation Approach**:

- Filter dwyl word list for exactly 5 letters, alphabetic only
- Use LLM to batch-generate short definitions (max 5 words each)
- Definitions must be paraphrased (no verbatim copies from existing sites)
- Use casual, non-formal wording
- Ensure exactly 1 definition per word
- Aim for 100-500 words minimum for MVP

### 2. Game Timer

**Goal**: Run a fixed 30-second countdown timer with optional auto-voice extension.

**Implementation**:

- Base duration: `GAME_DURATION_SECONDS = 30` (global constant)
- **Auto-voice mode**: +5 seconds per word when enabled (for screen reader users to have time to listen)
- Timer runs continuously (no pausing between words)
- Store timer state in React state
- Use `setInterval` with 1-second ticks
- Display remaining time in UI with color change when ≤5 seconds
- Emit "time up" event when timer reaches 0
- **Wrong answer penalty**: Deduct 1 second per incorrect submission
- **Auto-voice preference**: Persisted in localStorage across sessions

```typescript
const GAME_DURATION_SECONDS = 30;
const AUTO_VOICE_EXTRA_SECONDS = 5; // Extra time per word
const WRONG_ANSWER_PENALTY_SECONDS = 1;

interface TimerState {
  durationSeconds: number;
  isRunning: boolean;
  autoVoice: boolean; // Tracks if audio is enabled
}
```

### 3. Word Scrambling

**Goal**: Scramble word letters to maximize difficulty.

**Algorithm**:

1. Take original word
2. Perform 10 random shuffles
3. Calculate edit distance (Levenshtein) for each shuffle against original
4. Select shuffle with **highest** edit distance (hardest to unscramble)
5. If all shuffles match the original word, **skip the word** and select another

```typescript
function scrambleWord(word: string): string | null {
  const shuffles: string[] = [];
  for (let i = 0; i < 10; i++) {
    shuffles.push(shuffleLetters(word));
  }

  // Filter out any that match original
  const validShuffles = shuffles.filter((s) => s !== word);

  // If no valid shuffles, return null to skip this word
  if (validShuffles.length === 0) return null;

  // Sort by edit distance (descending) and pick first (hardest)
  return validShuffles.sort(
    (a, b) => editDistance(b, word) - editDistance(a, word),
  )[0];
}
```

### 4. Game Flow

**States**:

```typescript
type GameState =
  | { phase: "idle" }
  | {
      phase: "playing";
      currentWord: WordEntry;
      scrambled: string;
      autoVoice: boolean;
    }
  | { phase: "finished"; score: number; totalWords: number };
```

**Flow**:

1. **Idle**: Player sees "Start Game" button with optional "Enable auto text-to-speech" checkbox
   - Auto-voice preference is persisted in localStorage
   - Checkbox shows last selected state

2. **Playing**:
   - Timer starts (30 seconds base, +5 seconds per word if auto-voice enabled)
   - Timer runs continuously throughout session
   - Random word selected from bucket (no repeats in session)
   - Scrambled letters displayed with definition
   - If auto-voice enabled:
     - SpeechSynthesis automatically announces each letter
     - Scrambled letters shown for screen readers only
     - Extra 5 seconds added per word for listening time
   - Speaker button available for manual letter re-reading
   - Player types answer using on-screen keyboard or physical keyboard
   - If correct: increment score, load next word
   - If incorrect: deduct 1 second, show feedback, allow unlimited retries (no skip option)
   - Repeat until timer ends

3. **Finished**: Show final score + accuracy, display leaderboard from cookies, option to play again

### 5. Answer Validation

**Implementation**:

- Compare player input (case-insensitive, trimmed) to original word
- Return boolean result immediately

```typescript
function validateAnswer(input: string, expected: string): boolean {
  return input.trim().toLowerCase() === expected.toLowerCase();
}
```

### 6. Scoring & Storage

**MVP Approach**:

- Track score and attempts in React state during game
- Display both correct count AND accuracy (correct/attempts)
- **Leaderboard**: Store in cookie as JSON string, expires after 1 week
- Show leaderboard on results screen

```typescript
interface GameResult {
  id: string;
  timestamp: number; // Unix timestamp for JSON serialization
  score: number;
  totalAttempts: number;
  accuracy: number; // score / totalAttempts
}

// Cookie storage
const LEADERBOARD_COOKIE_NAME = "atoyr_leaderboard";
const LEADERBOARD_COOKIE_EXPIRY_DAYS = 7;
```

## UI Components

### Component Tree

```
App
├── StartScreen
│   ├── Game Instructions
│   ├── Auto-Voice Checkbox (localStorage-persisted)
│   └── StartButton
├── GameScreen
│   ├── Timer (color change when ≤5 seconds)
│   ├── ScrambledWord
│   ├── Definition
│   ├── SpeakerButton (manual text-to-speech)
│   ├── AnswerInput
│   ├── ScoreDisplay
│   └── OnScreenKeyboard (Wordle-style QWERTY)
└── ResultsScreen
    ├── FinalScore Display
    ├── AccuracyStats
    ├── Leaderboard (from cookies)
    └── PlayAgainButton
```

### Key UI Elements

1. **Start Screen**:
   - Title and tagline
   - Game instructions (rules, how to play)
   - Auto-voice checkbox with explanation
   - Start button

2. **Game Screen**:
   - **Timer Display**: Prominent countdown showing seconds remaining (color changes to red when ≤5 seconds)
   - **Score Display**: Shows correct count and accuracy percentage
   - **Scrambled Letters**: Large, clear display in individual boxes (Wordle-style)
   - **Definition**: Text hint showing the word's definition (max 5 words)
   - **Speaker Button**: Triggers SpeechSynthesis to read letters aloud
   - **Answer Display**: Shows current typed letters in boxes matching scrambled word length
   - **On-Screen Keyboard**: Wordle-like QWERTY layout at bottom with feedback states
   - **Feedback**: Visual indication of correct/incorrect answers

3. **Results Screen**:
   - Final score and accuracy metrics
   - Leaderboard showing all games from cookie (sorted, with timestamps)
   - Play again button
   - Link back to start screen

### Responsive Design (Mobile-First)

- **Max container width**: 500px (centered on larger screens)
- **Keyboard sizing**:
  - Optimized for 300px minimum width
  - Scales up with screen size
  - Maximum keyboard width: 430px
- Desktop browsers display at mobile viewport size (no wide layouts)

### Accessibility

- **SpeechSynthesis**: Automatically reads each letter when auto-voice is enabled (opt-in via checkbox)
- **Speaker button**: Manual trigger to re-read letters anytime
- **Keyboard support**: Physical keyboard input works alongside on-screen keyboard
- **Extra time for audio**: Auto-voice mode adds 5 seconds per word to allow time for listening
- **High contrast**: Dark theme with high contrast text and interactive elements
- **Screen reader friendly**: Proper semantic HTML and ARIA labels

## File Structure

```
packages/client/src/
├── data/
│   └── words.json           # Static word bucket
├── components/
│   ├── StartScreen.tsx
│   ├── GameScreen.tsx
│   ├── ResultsScreen.tsx
│   ├── Timer.tsx
│   └── ScrambledWord.tsx
├── hooks/
│   └── useGame.ts           # Game state management
├── utils/
│   ├── scramble.ts          # Scrambling logic
│   ├── editDistance.ts      # Levenshtein distance
│   └── validation.ts        # Answer validation
├── types/
│   └── game.ts              # TypeScript interfaces
├── App.tsx
└── main.tsx

packages/shared/src/
└── index.ts                 # Shared types (for future use)

scripts/
└── scrape-words.ts          # Word scraper script
```

## Testing Strategy

### Unit Tests (Vitest)

1. **scramble.ts**
   - Returns different string than input
   - Returns string with same letters
   - Handles edge cases (all same letters)

2. **editDistance.ts**
   - Correct distance calculations
   - Edge cases (empty strings, identical strings)

3. **validation.ts**
   - Case insensitivity
   - Whitespace handling
   - Exact match required

4. **useGame.ts**
   - State transitions
   - Score incrementing
   - Timer behavior

## Implementation Status

### Completed ✅

1. [x] Create word scraper script
2. [x] Generate initial word bucket JSON
3. [x] Implement `editDistance` utility with tests
4. [x] Implement `scramble` utility with tests
5. [x] Implement `validation` utility with tests
6. [x] Create `useGame` hook with tests
7. [x] Build UI components (StartScreen, GameScreen, ResultsScreen)
8. [x] Implement on-screen Wordle-style keyboard
9. [x] Add auto-voice feature with localStorage persistence
10. [x] Implement leaderboard with cookie storage (7-day expiry)
11. [x] Add speaker button for manual text-to-speech
12. [x] Implement responsive mobile-first design
13. [x] Add dark theme styling
14. [x] Integration testing and polish
15. [x] Deploy MVP

---

## Implementation Notes

### Auto-Voice Feature

The auto-voice feature is designed for accessibility and screen reader users:

- **Optional Opt-In**: Players can enable/disable via checkbox on start screen
- **Persistent Preference**: Last selection stored in localStorage
- **Extra Time Allowance**: +5 seconds per word when enabled to allow listening time
- **Letter Pronunciation**: Browser's SpeechSynthesis API reads each letter individually
- **Manual Re-read**: Speaker button available anytime during gameplay
- **Screen Reader Compatibility**: When enabled, scrambled letters are shown for screen readers

### Leaderboard Storage

- Stored in browser cookies with 7-day expiration
- JSON format: array of GameResult objects
- Includes timestamp, score, total attempts, and calculated accuracy
- Auto-clears after 7 days of inactivity
- Per-browser storage (not synced across devices)

### Game State Management

- Uses React hooks for centralized state
- Set-based tracking prevents word repeats within session
- Timer runs independently via setInterval
- Score automatically incremented on correct answers
- Time penalties applied immediately on incorrect answers

---

## Resolved Questions

All questions have been resolved and incorporated into the spec above. Keeping for historical reference:

<details>
<summary>Click to expand resolved questions</summary>

1. **Word Source**: ✅ Use dwyl/english-words + LLM-generated definitions (paraphrased, non-formal, max 5 words)

2. **Scrambling Algorithm**: ✅ Use highest edit distance (hardest). Skip word if all shuffles match original.

3. **Definition Display**: ✅ Max 5 words, 1 definition per word, non-formal wording.

4. **Wrong Answer Handling**: ✅ 1 second penalty per wrong answer. No skip option. Unlimited retries.

5. **Timer Configuration**: ✅ Fixed 30 seconds base (global constant) + 5 seconds per word if auto-voice enabled. Runs continuously.

6. **Score Display**: ✅ Show correct count AND accuracy. Leaderboard stored in cookie (7 day expiry).

7. **Word Repetition**: ✅ No repeats within session. Cross-session tracking deferred to full version.

8. **Accessibility**: ✅ Optional auto-voice via checkbox with localStorage persistence. +5 seconds per word. Manual speaker button. Mobile-first Wordle-style keyboard. Max width 430px.

9. **Auto-Voice Extra Time**: ✅ +5 seconds per word to allow listening time for screen readers and accessibility users.

</details>
