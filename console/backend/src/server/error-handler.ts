import type { NextFunction, Request, Response } from 'express';
import { HttpError } from './http-error.js';

export function errorHandler(error: unknown, _request: Request, response: Response, _next: NextFunction) {
  if (error instanceof HttpError) {
    response.status(error.status).json({ message: error.message });
    return;
  }

  const message = error instanceof Error ? error.message : 'internal server error';
  response.status(500).json({ message });
}
