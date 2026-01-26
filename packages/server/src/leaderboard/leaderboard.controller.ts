import { leaderboardQuerySchema } from '@atoyr/shared';
import { Controller, Get, Query, UseFilters } from '@nestjs/common';
import { HttpExceptionFilter } from '../common/filters/http-exception.filter';
import { LeaderboardService } from './leaderboard.service';

@Controller('api/leaderboard')
@UseFilters(HttpExceptionFilter)
export class LeaderboardController {
  constructor(private leaderboardService: LeaderboardService) {}

  @Get()
  async getLeaderboard(@Query('limit') limit?: string, @Query('offset') offset?: string) {
    const validated = leaderboardQuerySchema.parse({
      limit: limit ? parseInt(limit, 10) : undefined,
      offset: offset ? parseInt(offset, 10) : undefined,
    });

    return this.leaderboardService.getLeaderboard(validated.limit, validated.offset);
  }
}
