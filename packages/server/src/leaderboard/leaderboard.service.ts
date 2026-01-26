import { LeaderboardEntry, LeaderboardResponse } from '@atoyr/shared';
import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { ResultEntity } from './result.entity';

@Injectable()
export class LeaderboardService {
  constructor(
    @InjectRepository(ResultEntity)
    private resultRepository: Repository<ResultEntity>,
  ) {}

  async getLeaderboard(limit: number = 10, offset: number = 0): Promise<LeaderboardResponse> {
    // Cap limit at 100
    const safeLimit = Math.min(Math.max(limit, 1), 100);
    const safeOffset = Math.max(offset, 0);

    const [results, total] = await this.resultRepository.findAndCount({
      order: { score: 'DESC', timestamp: 'DESC' },
      take: safeLimit,
      skip: safeOffset,
    });

    const entries: LeaderboardEntry[] = results.map((result: ResultEntity, index: number) => ({
      id: result.id,
      rank: safeOffset + index + 1,
      score: result.score,
      accuracy: result.accuracy,
      timestamp: result.timestamp,
    }));

    return {
      entries,
      totalGames: total,
      lastUpdated: Date.now(),
    };
  }

  async getTopScores(limit: number = 10): Promise<LeaderboardEntry[]> {
    const results = await this.resultRepository.find({
      order: { score: 'DESC', timestamp: 'DESC' },
      take: limit,
    });

    return results.map((result: ResultEntity, index: number) => ({
      id: result.id,
      rank: index + 1,
      score: result.score,
      accuracy: result.accuracy,
      timestamp: result.timestamp,
    }));
  }
}
