import { StartGameRequest, SubmitAnswerRequest, startGameRequestSchema, submitAnswerRequestSchema } from '@atoyr/shared';
import { BadRequestException, Body, Controller, HttpCode, HttpStatus, Post, UseFilters } from '@nestjs/common';
import { HttpExceptionFilter } from '../common/filters/http-exception.filter';
import { GameService } from './game.service';

@Controller('api/game')
@UseFilters(HttpExceptionFilter)
export class GameController {
  constructor(private gameService: GameService) {}

  @Post('start')
  @HttpCode(HttpStatus.CREATED)
  async startGame(@Body() body: StartGameRequest) {
    const validated = startGameRequestSchema.parse(body);
    return this.gameService.startGame(validated.autoVoice);
  }

  @Post('start/test')
  @HttpCode(HttpStatus.CREATED)
  async startGameTest(@Body() body: StartGameRequest & { durationSeconds?: number }) {
    if (process.env.NODE_ENV === 'production') {
      throw new BadRequestException('Test endpoint not available in production');
    }

    const validated = startGameRequestSchema.parse(body);
    return this.gameService.startGame(validated.autoVoice, body.durationSeconds);
  }

  @Post('answer')
  @HttpCode(HttpStatus.OK)
  async submitAnswer(@Body() body: SubmitAnswerRequest) {
    try {
      const validated = submitAnswerRequestSchema.parse(body);
      return this.gameService.submitAnswer(validated.sessionId, validated.token, validated.answer);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      if (errorMessage.includes('SESSION_NOT_FOUND')) {
        throw new BadRequestException('Game session not found');
      }
      if (errorMessage.includes('INVALID_TOKEN')) {
        throw new BadRequestException('Invalid word token');
      }
      if (errorMessage.includes('GAME_NOT_STARTED')) {
        throw new BadRequestException('Game has not been started');
      }
      throw error;
    }
  }
}
