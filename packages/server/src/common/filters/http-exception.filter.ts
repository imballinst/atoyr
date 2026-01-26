import { ArgumentsHost, Catch, ExceptionFilter, HttpException, HttpStatus } from '@nestjs/common';
import { Response } from 'express';
import { ZodError } from 'zod';

@Catch()
export class HttpExceptionFilter implements ExceptionFilter {
  catch(exception: unknown, host: ArgumentsHost) {
    const ctx = host.switchToHttp();
    const response = ctx.getResponse<Response>();
    const request = ctx.getRequest();

    let status = HttpStatus.INTERNAL_SERVER_ERROR;
    let message = 'Internal server error';
    let code = 'INTERNAL_ERROR';

    if (exception instanceof HttpException) {
      status = exception.getStatus();
      const exceptionResponse = exception.getResponse();
      message = (exceptionResponse as any).message || exception.message;
      code = (exceptionResponse as any).code || 'HTTP_ERROR';
    } else if (exception instanceof ZodError) {
      status = HttpStatus.BAD_REQUEST;
      message = exception.issues[0]?.message || 'Invalid request data';
      code = 'INVALID_REQUEST';
    } else if (exception instanceof Error) {
      message = exception.message;
      code = 'DATABASE_ERROR';
      // Log full error in development
      if (process.env.NODE_ENV !== 'production') {
        console.error('Error:', exception);
      }
    }

    response.status(status).json({
      statusCode: status,
      error: HttpStatus[status] || 'Error',
      message,
      code,
      path: request.url,
      timestamp: new Date().toISOString(),
    });
  }
}
