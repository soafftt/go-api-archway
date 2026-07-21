import type { NextFunction, Request, RequestHandler, Response } from 'express';
import { Router } from 'express';
import { ZodError } from 'zod';
import { ConflictError, NotFoundError, UpstreamAdminService } from '../application/upstream-admin-service.js';
import { HttpError } from './http-error.js';

export function createUpstreamRoutes(service: UpstreamAdminService) {
  const router = Router();

  router.post('/preview-match', asyncHandler(async (request, response) => {
    const gatewayPath = String(request.body?.gatewayPath ?? '');
    const method = String(request.body?.method ?? 'GET');
    response.json(await service.preview(gatewayPath, method));
  }));

  router.post('/:serviceName/republish', asyncHandler(async (request, response) => {
    await service.republish(request.params.serviceName);
    response.status(202).json({ message: 'republish queued' });
  }));

  router.get('/', asyncHandler(async (_request, response) => {
    response.json(await service.list());
  }));

  router.get('/:serviceName', asyncHandler(async (request, response) => {
    response.json(await service.get(request.params.serviceName));
  }));

  router.post('/', asyncHandler(async (request, response) => {
    response.status(201).json(await service.create(request.body));
  }));

  router.put('/:serviceName', asyncHandler(async (request, response) => {
    response.json(await service.update(request.params.serviceName, request.body));
  }));

  router.put('/:serviceName/resources/upsert', asyncHandler(async (request, response) => {
    response.json(await service.upsertResource(request.params.serviceName, request.body));
  }));

  router.delete('/:serviceName', asyncHandler(async (request, response) => {
    await service.delete(request.params.serviceName);
    response.status(204).send();
  }));

  return router;
}

function asyncHandler(
  handler: (request: Request, response: Response, next: NextFunction) => Promise<void>,
): RequestHandler {
  return async (request, response, next) => {
    try {
      await handler(request, response, next);
    } catch (error) {
      if (error instanceof ZodError) {
        next(new HttpError(400, error.issues.map((issue) => `${issue.path.join('.')}: ${issue.message}`).join(', ')));
        return;
      }
      if (error instanceof ConflictError) {
        next(new HttpError(409, error.message));
        return;
      }
      if (error instanceof NotFoundError) {
        next(new HttpError(404, error.message));
        return;
      }
      next(error);
    }
  };
}
