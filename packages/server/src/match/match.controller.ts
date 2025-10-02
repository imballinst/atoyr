import { Body, Controller, Get, Post, Param, Res } from '@nestjs/common';
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

  @Get('events/:playerId')
  events(@Param('playerId') playerId: string, @Res() res: any) {
    // Set SSE headers
    res.setHeader('Content-Type', 'text/event-stream');
    res.setHeader('Cache-Control', 'no-cache');
    res.setHeader('Connection', 'keep-alive');
    res.setHeader('X-Accel-Buffering', 'no');
    res.flushHeaders && res.flushHeaders();
    res.write(': connected\n\n');

    this.matchService.addSseClient(playerId, res);

    reqCleanup();

    function reqCleanup() {
      res.on('close', () => {
        try {
          res.end();
        } catch (e) {}
        try {
          // Remove from service
          (res as any).locals &&
            (res as any).locals.matchService &&
            (res as any).locals.matchService.removeSseClient(playerId, res);
        } catch (e) {}
      });
    }
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
