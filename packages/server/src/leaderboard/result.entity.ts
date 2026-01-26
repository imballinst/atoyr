import { Column, CreateDateColumn, Entity, Index, PrimaryColumn } from 'typeorm';

@Entity('results')
@Index('idx_results_score', ['score'], { order: 'DESC' })
@Index('idx_results_timestamp', ['timestamp'], { order: 'DESC' })
export class ResultEntity {
  @PrimaryColumn('text')
  id: string;

  @Column('text')
  sessionId: string;

  @CreateDateColumn({ type: 'integer' })
  timestamp: number;

  @Column('integer')
  score: number;

  @Column('integer')
  totalAttempts: number;

  @Column('real')
  accuracy: number;

  @Column('integer')
  durationMs: number;
}
