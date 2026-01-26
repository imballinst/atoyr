import { AUTO_VOICE_EXTRA_SECONDS, GAME_DURATION_SECONDS, scrambleWord, validateAnswer, WRONG_ANSWER_PENALTY_SECONDS } from '@atoyr/shared';
import { Injectable } from '@nestjs/common';
import { EventEmitter2 } from '@nestjs/event-emitter';
import { v4 as uuidv4 } from 'uuid';
import { SessionService } from '../session/session.service';
import { WordService } from '../word/word.service';

@Injectable()
export class GameService {
  private activeTimers = new Map<string, NodeJS.Timeout>();

  constructor(
    private sessionService: SessionService,
    private wordService: WordService,
    private eventEmitter: EventEmitter2,
  ) {}

  async startGame(autoVoice: boolean, durationSeconds?: number): Promise<{ sessionId: string; expiresAt: number }> {
    const sessionId = uuidv4();
    const now = Date.now();
    const initialSeconds = durationSeconds ?? (autoVoice ? GAME_DURATION_SECONDS + AUTO_VOICE_EXTRA_SECONDS : GAME_DURATION_SECONDS);
    const expiresAt = now + (initialSeconds + 600) * 1000; // Duration + 10 min buffer

    await this.sessionService.create({
      id: sessionId,
      createdAt: now,
      expiresAt,
      state: {
        phase: 'playing',
        score: 0,
        totalAttempts: 0,
        remainingSeconds: initialSeconds,
        usedWords: [],
        currentWordToken: null,
      },
      autoVoice,
    });

    return { sessionId, expiresAt };
  }

  async initializeSSE(sessionId: string): Promise<void> {
    const session = await this.sessionService.findById(sessionId);
    if (!session) throw new Error('SESSION_NOT_FOUND');

    // Emit session started
    this.eventEmitter.emit(`sse.${sessionId}`, {
      type: 'session:started',
      sessionId,
      autoVoice: session.autoVoice,
      remainingSeconds: session.state.remainingSeconds,
      timestamp: Date.now(),
    });

    // Select and emit first word
    await this.emitNextWord(sessionId);

    // Start timer
    this.startTimer(sessionId);
  }

  private async emitNextWord(sessionId: string): Promise<void> {
    const session = await this.sessionService.findById(sessionId);
    if (!session || session.state.phase !== 'playing') return;

    const word = await this.wordService.getRandomWord(session.state.usedWords);
    if (!word) {
      await this.finishGame(sessionId);
      return;
    }

    const scrambled = scrambleWord(word.word);
    if (!scrambled) {
      // Skip word, try another
      session.state.usedWords.push(word.word);
      await this.sessionService.update(sessionId, session);
      return this.emitNextWord(sessionId);
    }

    const token = uuidv4();
    session.state.usedWords.push(word.word);
    session.state.currentWordToken = token;
    await this.sessionService.update(sessionId, session);
    await this.sessionService.setCurrentWord(sessionId, word.word, token);

    this.eventEmitter.emit(`sse.${sessionId}`, {
      type: 'word:new',
      scrambled: scrambled.toUpperCase(),
      definition: word.definition,
      token,
      wordIndex: session.state.usedWords.length,
      timestamp: Date.now(),
    });

    // Add extra time for auto-voice
    if (session.autoVoice) {
      session.state.remainingSeconds += AUTO_VOICE_EXTRA_SECONDS;
      await this.sessionService.update(sessionId, session);
    }
  }

  private startTimer(sessionId: string): void {
    const timer = setInterval(async () => {
      const session = await this.sessionService.findById(sessionId);
      if (!session || session.state.phase !== 'playing') {
        clearInterval(timer);
        this.activeTimers.delete(sessionId);
        return;
      }

      session.state.remainingSeconds -= 1;

      if (session.state.remainingSeconds <= 0) {
        await this.finishGame(sessionId);
        clearInterval(timer);
        this.activeTimers.delete(sessionId);
        return;
      }

      await this.sessionService.update(sessionId, session);

      this.eventEmitter.emit(`sse.${sessionId}`, {
        type: 'timer:tick',
        remainingSeconds: session.state.remainingSeconds,
        timestamp: Date.now(),
      });
    }, 1000);

    this.activeTimers.set(sessionId, timer);
  }

  async submitAnswer(
    sessionId: string,
    token: string,
    answer: string,
  ): Promise<{
    correct: boolean;
    score: number;
    totalAttempts: number;
    remainingSeconds: number;
    gameOver: boolean;
  }> {
    const session = await this.sessionService.findById(sessionId);
    if (!session) throw new Error('SESSION_NOT_FOUND');
    if (session.state.phase !== 'playing') throw new Error('GAME_NOT_STARTED');
    if (session.state.currentWordToken !== token) throw new Error('INVALID_TOKEN');

    const currentWord = await this.sessionService.getCurrentWord(sessionId);
    if (!currentWord) throw new Error('GAME_NOT_STARTED');

    const isCorrect = validateAnswer(answer, currentWord);
    session.state.totalAttempts += 1;

    if (isCorrect) {
      session.state.score += 1;
      await this.sessionService.update(sessionId, session);
      await this.emitNextWord(sessionId);
    } else {
      const penalty = Math.min(WRONG_ANSWER_PENALTY_SECONDS, session.state.remainingSeconds);
      session.state.remainingSeconds -= penalty;
      await this.sessionService.update(sessionId, session);

      this.eventEmitter.emit(`sse.${sessionId}`, {
        type: 'timer:penalty',
        penaltySeconds: penalty,
        remainingSeconds: session.state.remainingSeconds,
        timestamp: Date.now(),
      });

      if (session.state.remainingSeconds <= 0) {
        await this.finishGame(sessionId);
      }
    }

    const updatedSession = await this.sessionService.findById(sessionId);
    return {
      correct: isCorrect,
      score: updatedSession!.state.score,
      totalAttempts: updatedSession!.state.totalAttempts,
      remainingSeconds: updatedSession!.state.remainingSeconds,
      gameOver: updatedSession!.state.phase === 'finished',
    };
  }

  private async finishGame(sessionId: string): Promise<void> {
    const session = await this.sessionService.findById(sessionId);
    if (!session) return;

    session.state.phase = 'finished';
    session.state.remainingSeconds = 0;
    await this.sessionService.update(sessionId, session);

    const accuracy = session.state.totalAttempts > 0 ? session.state.score / session.state.totalAttempts : 0;

    const resultId = await this.sessionService.saveResult(sessionId, {
      score: session.state.score,
      totalAttempts: session.state.totalAttempts,
      accuracy,
      durationMs: Date.now() - session.createdAt,
    });

    this.eventEmitter.emit(`sse.${sessionId}`, {
      type: 'game:finished',
      score: session.state.score,
      totalAttempts: session.state.totalAttempts,
      accuracy,
      resultId,
      timestamp: Date.now(),
    });

    // Cleanup timer
    const timer = this.activeTimers.get(sessionId);
    if (timer) {
      clearInterval(timer);
      this.activeTimers.delete(sessionId);
    }
  }

  cleanupSession(sessionId: string): void {
    const timer = this.activeTimers.get(sessionId);
    if (timer) {
      clearInterval(timer);
      this.activeTimers.delete(sessionId);
    }
  }
}
