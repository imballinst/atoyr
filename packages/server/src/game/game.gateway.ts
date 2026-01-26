import { SSEEvent, uuidSchema } from '@atoyr/shared';
import { Controller, Get, HttpException, HttpStatus, Param, Res, UseFilters } from '@nestjs/common';
import { EventEmitter2 } from '@nestjs/event-emitter';
import { Response } from 'express';
import { HttpExceptionFilter } from '../common/filters/http-exception.filter';
import { SessionService } from '../session/session.service';
import { GameService } from './game.service';

@Controller('api/game')
@UseFilters(HttpExceptionFilter)
export class GameGateway {
  constructor(
    private gameService: GameService,
    private sessionService: SessionService,
    private eventEmitter: EventEmitter2,
  ) {}

  @Get('sse/:sessionId')
  async stream(@Param('sessionId') sessionId: string, @Res() res: Response) {
    try {
      // Validate UUID format
      uuidSchema.parse(sessionId);

      const session = await this.sessionService.findById(sessionId);

      if (!session) {
        throw new HttpException('Session not found', HttpStatus.NOT_FOUND);
      }

      if (session.expiresAt < Date.now()) {
        throw new HttpException('Session expired', HttpStatus.GONE);
      }

      // Set SSE headers
      res.setHeader('Content-Type', 'text/event-stream');
      res.setHeader('Cache-Control', 'no-cache');
      res.setHeader('Connection', 'keep-alive');
      res.setHeader('X-Accel-Buffering', 'no'); // Disable nginx buffering
      res.flushHeaders();

      // Event listener for this session
      const eventHandler = (event: SSEEvent) => {
        if (!res.destroyed) {
          res.write(`event: ${event.type}\n`);
          res.write(`data: ${JSON.stringify(event)}\n\n`);
        }
      };

      this.eventEmitter.on(`sse.${sessionId}`, eventHandler);

      // Initialize game and start emitting events
      try {
        await this.gameService.initializeSSE(sessionId);
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : String(error);
        if (!res.destroyed) {
          res.write(`event: error\n`);
          res.write(
            `data: ${JSON.stringify({
              type: 'error',
              message: errorMessage,
              timestamp: Date.now(),
            })}\n\n`,
          );
          res.end();
        }
        this.eventEmitter.off(`sse.${sessionId}`, eventHandler);
        return;
      }

      // Handle client disconnect
      res.on('close', () => {
        this.eventEmitter.off(`sse.${sessionId}`, eventHandler);
        this.gameService.cleanupSession(sessionId);
      });
    } catch (error) {
      if (error instanceof HttpException) {
        throw error;
      }
      throw new HttpException('Bad request', HttpStatus.BAD_REQUEST);
    }
  }
}
