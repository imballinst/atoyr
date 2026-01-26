import { GameSession } from '@atoyr/shared';
import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { v4 as uuidv4 } from 'uuid';
import { ResultEntity } from '../leaderboard/result.entity';
import { SessionEntity } from './session.entity';

@Injectable()
export class SessionService {
  constructor(
    @InjectRepository(SessionEntity)
    private sessionRepository: Repository<SessionEntity>,
    @InjectRepository(ResultEntity)
    private resultRepository: Repository<ResultEntity>,
  ) {}

  async create(session: GameSession): Promise<void> {
    const entity = new SessionEntity();
    entity.id = session.id;
    entity.createdAt = session.createdAt;
    entity.expiresAt = session.expiresAt;
    entity.autoVoice = session.autoVoice ? 1 : 0;
    entity.setState(session.state);

    await this.sessionRepository.save(entity);
  }

  async findById(id: string): Promise<GameSession | null> {
    const entity = await this.sessionRepository.findOne({ where: { id } });
    if (!entity) return null;

    return {
      id: entity.id,
      createdAt: entity.createdAt,
      expiresAt: entity.expiresAt,
      state: entity.getState(),
      autoVoice: entity.autoVoice === 1,
    };
  }

  async update(id: string, session: GameSession): Promise<void> {
    const entity = await this.sessionRepository.findOne({ where: { id } });
    if (!entity) throw new Error('SESSION_NOT_FOUND');

    entity.setState(session.state);
    entity.expiresAt = session.expiresAt;
    entity.autoVoice = session.autoVoice ? 1 : 0;

    await this.sessionRepository.save(entity);
  }

  async setCurrentWord(sessionId: string, word: string, token: string): Promise<void> {
    await this.sessionRepository.update(
      { id: sessionId },
      {
        currentWord: word,
        currentWordToken: token,
      },
    );
  }

  async getCurrentWord(sessionId: string): Promise<string | null> {
    const entity = await this.sessionRepository.findOne({
      where: { id: sessionId },
      select: ['currentWord'],
    });
    return entity?.currentWord || null;
  }

  async saveResult(
    sessionId: string,
    result: {
      score: number;
      totalAttempts: number;
      accuracy: number;
      durationMs: number;
    },
  ): Promise<string> {
    const resultId = uuidv4();
    const resultEntity = new ResultEntity();
    resultEntity.id = resultId;
    resultEntity.sessionId = sessionId;
    resultEntity.timestamp = Date.now();
    resultEntity.score = result.score;
    resultEntity.totalAttempts = result.totalAttempts;
    resultEntity.accuracy = result.accuracy;
    resultEntity.durationMs = result.durationMs;

    await this.resultRepository.save(resultEntity);
    return resultId;
  }

  async cleanupExpiredSessions(): Promise<void> {
    const now = Date.now();
    await this.sessionRepository.createQueryBuilder().delete().where('expiresAt < :now', { now }).execute();
  }
}
