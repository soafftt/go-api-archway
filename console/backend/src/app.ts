import cors from 'cors';
import express from 'express';
import { UpstreamAdminService } from './application/upstream-admin-service.js';
import { errorHandler } from './server/error-handler.js';
import { createUpstreamRoutes } from './server/upstream-routes.js';

export function createApp(service: UpstreamAdminService) {
  const app = express();

  app.use(cors());
  app.use(express.json({ limit: '2mb' }));

  app.get('/health', (_request, response) => {
    response.json({ status: 'ok' });
  });

  app.use('/api/v1/upstream-services', createUpstreamRoutes(service));
  app.use(errorHandler);

  return app;
}
