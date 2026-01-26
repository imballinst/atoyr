import { Module } from '@nestjs/common';
import { SessionModule } from '../session/session.module';
import { WordModule } from '../word/word.module';
import { GameController } from './game.controller';
import { GameGateway } from './game.gateway';
import { GameService } from './game.service';

@Module({
  imports: [SessionModule, WordModule],
  controllers: [GameController, GameGateway],
  providers: [GameService],
  exports: [GameService],
})
export class GameModule {}
