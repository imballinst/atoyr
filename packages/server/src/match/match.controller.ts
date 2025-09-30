import { Body, Controller, Get, Post } from '@nestjs/common';
import { MatchService } from './match.service';

@Controller()
export class MatchController {
  constructor(private readonly matchService: MatchService) {}

  @Get()
  getRoot() {
    return this.matchService.root();
  }

  @Post('player')
  createPlayer(@Body() body: { id: string; name?: string; level?: number }) {
    const { id, name, level } = body;
    return this.matchService.createPlayer(id, name, level);
  }

  @Get('player/:id')
  getPlayer(@Body() body: { id: string }) {
    return this.matchService.getPlayer(body.id);
  }

  @Post('matchmake')
  matchmake(@Body() body: { id: string; opponentId?: string }) {
    const { id, opponentId } = body;
    return this.matchService.matchmake(id, opponentId);
  }

  @Get('match/:id')
  getMatch(@Body() body: { id: string }) {
    return this.matchService.getMatch(body.id);
  }

  @Post('resolve')
  resolve(
    @Body() body: { matchId: string; attackerId: string; attackerTime?: number; defenderTime?: number },
  ) {
    return this.matchService.resolveTurn(body.matchId, body.attackerId, body.attackerTime, body.defenderTime);
  }
}
