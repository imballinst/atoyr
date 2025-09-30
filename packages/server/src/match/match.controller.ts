import { Body, Controller, Get, Post, Param, Sse } from '@nestjs/common';
import { MatchService } from './match.service';
import { Observable } from 'rxjs';
import { filter, map } from 'rxjs/operators';
import type { MessageEvent } from '@nestjs/common';

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
  getPlayer(@Param('id') id: string) {
    return this.matchService.getPlayer(id);
  }

  @Post('matchmake')
  matchmake(@Body() body: { id: string; opponentId?: string }) {
    const { id, opponentId } = body;
    return this.matchService.matchmake(id, opponentId);
  }

  @Get('match/:id')
  getMatch(@Param('id') id: string) {
    return this.matchService.getMatch(id);
  }

  @Sse('events/:playerId')
  events(@Param('playerId') playerId: string): Observable<MessageEvent> {
    return this.matchService.eventStream.pipe(
      filter((e) => !e.playerId || e.playerId === playerId),
      map((e) => ({ event: e.event, data: e.data }) as MessageEvent),
    );
  }

  @Post('match/:id/join')
  joinMatch(@Param('id') id: string, @Body() body: { playerId: string }) {
    const { playerId } = body;
    return this.matchService.joinMatch(id, playerId);
  }

  @Post('resolve')
  resolve(
    @Body()
    body: {
      matchId: string;
      attackerId: string;
      attackerTime?: number;
      defenderTime?: number;
    },
  ) {
    return this.matchService.resolveTurn(
      body.matchId,
      body.attackerId,
      body.attackerTime,
      body.defenderTime,
    );
  }
}
