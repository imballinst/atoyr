import { Body, Controller, Get, Post } from '@nestjs/common';
import { MatchService } from './match.service';

@Controller()
export class MatchController {
  constructor(private readonly matchService: MatchService) {}

  @Get()
  getRoot() {
    return this.matchService.root();
  }

  @Post('matchmake')
  matchmake() {
    return this.matchService.matchmake();
  }

  @Post('resolve')
  resolve(@Body() body: { attackerTime?: number; defenderTime?: number }) {
    const { attackerTime, defenderTime } = body;
    return this.matchService.resolve(attackerTime, defenderTime);
  }
}
