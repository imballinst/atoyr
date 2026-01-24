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

**Goal**: Run a fixed 30-second countdown timer.

**Implementation**:

- Fixed duration: `GAME_DURATION_SECONDS = 30` (global constant)
- Timer runs continuously (no pausing between words)
- Store timer state in React state
- Use `setInterval` with 1-second ticks
- Display remaining time in UI
- Emit "time up" event when timer reaches 0
- **Wrong answer penalty**: Deduct 1 second per incorrect submission

```typescript
const GAME_DURATION_SECONDS = 30;
const WRONG_ANSWER_PENALTY_SECONDS = 1;

interface TimerState {
  remainingSeconds: number;
  isRunning: boolean;
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
  return validShuffles.sort((a, b) => editDistance(b, word) - editDistance(a, word))[0];
}
```

### 4. Game Flow

**States**:

```typescript
type GameState =
  | { phase: 'idle' }
  | { phase: 'playing'; currentWord: WordEntry; scrambled: string }
  | { phase: 'finished'; score: number; totalWords: number };
```

**Flow**:

1. **Idle**: Player sees "Start Game" button
2. **Playing**:
   - Timer starts (30 seconds, runs continuously)
   - Random word selected from bucket (no repeats in session)
   - Scrambled letters displayed with definition
   - SpeechSynthesis announces each letter
   - Player types answer using on-screen keyboard or physical keyboard
   - If correct: increment score, load next word
   - If incorrect: deduct 1 second, show feedback, allow unlimited retries (no skip option)
   - Repeat until timer ends
3. **Finished**: Show final score + accuracy, save to leaderboard, option to play again

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
const LEADERBOARD_COOKIE_NAME = 'atoyr_leaderboard';
const LEADERBOARD_COOKIE_EXPIRY_DAYS = 7;
```

## UI Components

### Component Tree

```
App
├── StartScreen
│   └── StartButton
├── GameScreen
│   ├── Timer
│   ├── ScrambledWord
│   ├── Definition
│   ├── AnswerInput
│   └── Score
└── ResultsScreen
    ├── FinalScore
    └── PlayAgainButton
```

### Key UI Elements

1. **Timer Display**: Prominent countdown showing seconds remaining
2. **Scrambled Letters**: Large, clear display of scrambled letters in individual boxes (Wordle-style)
3. **Definition**: Text hint showing the word's definition (max 5 words)
4. **Speaker Button**: Triggers SpeechSynthesis to read letters aloud
5. **Answer Display**: Shows current typed letters in boxes
6. **On-Screen Keyboard**: Wordle-like QWERTY keyboard at bottom
7. **Score Counter**: Shows correct count and accuracy

### Responsive Design (Mobile-First)

- **Max container width**: 500px (centered on larger screens)
- **Keyboard sizing**:
  - Optimized for 300px minimum width
  - Scales up with screen size
  - Maximum keyboard width: 430px
- Desktop browsers display at mobile viewport size (no wide layouts)

### Accessibility

- **SpeechSynthesis**: Automatically reads each letter when a new word appears
- **Speaker button**: Manual trigger to re-read letters
- **Keyboard support**: Physical keyboard input works alongside on-screen keyboard

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

## Implementation Order

1. [ ] Create word scraper script
2. [ ] Generate initial word bucket JSON
3. [ ] Implement `editDistance` utility with tests
4. [ ] Implement `scramble` utility with tests
5. [ ] Implement `validation` utility with tests
6. [ ] Create `useGame` hook with tests
7. [ ] Build UI components
8. [ ] Integration testing
9. [ ] Polish and styling

---

## Resolved Questions

All questions have been resolved and incorporated into the spec above. Keeping for historical reference:

<details>
<summary>Click to expand resolved questions</summary>

1. **Word Source**: ✅ Use dwyl/english-words + LLM-generated definitions (paraphrased, non-formal, max 5 words)

2. **Scrambling Algorithm**: ✅ Use highest edit distance (hardest). Skip word if all shuffles match original.

3. **Definition Display**: ✅ Max 5 words, 1 definition per word, non-formal wording.

4. **Wrong Answer Handling**: ✅ 1 second penalty per wrong answer. No skip option. Unlimited retries.

5. **Timer Configuration**: ✅ Fixed 30 seconds (global constant). Runs continuously.

6. **Score Display**: ✅ Show correct count AND accuracy. Leaderboard stored in cookie (1 week expiry).

7. **Word Repetition**: ✅ No repeats within session. Cross-session tracking deferred to full version.

8. **Accessibility**: ✅ SpeechSynthesis for letters. Mobile-first Wordle-style keyboard. Max width 500px, keyboard max 430px.

</details>
