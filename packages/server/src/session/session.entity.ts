import { GameSessionState } from '@atoyr/shared';
import { Column, CreateDateColumn, Entity, Index, PrimaryColumn } from 'typeorm';

@Entity('sessions')
@Index('idx_sessions_expires', ['expiresAt'])
export class SessionEntity {
  @PrimaryColumn('text')
  id!: string;

  @CreateDateColumn({ type: 'integer' })
  createdAt!: number;

  @Column('integer')
  expiresAt!: number;

  @Column('text', { default: 'idle' })
  phase!: 'idle' | 'playing' | 'finished';

  @Column('integer', { default: 0 })
  score!: number;

  @Column('integer', { default: 0 })
  totalAttempts!: number;

  @Column('integer')
  remainingSeconds!: number;

  @Column('integer', { default: 0 })
  autoVoice!: number; // Store boolean as integer for SQLite

  @Column('text', { default: '[]' })
  usedWords!: string; // JSON array

  @Column('text', { nullable: true })
  currentWord!: string | null;

  @Column('text', { nullable: true })
  currentWordToken!: string | null;

  // Helper methods for state management
  getState(): GameSessionState {
    return {
      phase: this.phase,
      score: this.score,
      totalAttempts: this.totalAttempts,
      remainingSeconds: this.remainingSeconds,
      usedWords: JSON.parse(this.usedWords),
      currentWordToken: this.currentWordToken,
    };
  }

  setState(state: GameSessionState): void {
    this.phase = state.phase;
    this.score = state.score;
    this.totalAttempts = state.totalAttempts;
    this.remainingSeconds = state.remainingSeconds;
    this.usedWords = JSON.stringify(state.usedWords);
    this.currentWordToken = state.currentWordToken;
  }
}
