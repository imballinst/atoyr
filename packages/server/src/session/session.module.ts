import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { ResultEntity } from '../leaderboard/result.entity';
import { SessionEntity } from './session.entity';
import { SessionService } from './session.service';

@Module({
  imports: [TypeOrmModule.forFeature([SessionEntity, ResultEntity])],
  providers: [SessionService],
  exports: [SessionService],
})
export class SessionModule {}
