import { Redis } from 'ioredis';
import { createApp } from './app.js';
import { UpstreamAdminService } from './application/upstream-admin-service.js';
import { loadConfig } from './config/env.js';
import { toGatewaySnapshotJson } from './domain/upstream.js';
import { createPostgresPool } from './infra/postgres/postgres.js';
import { PostgresRouteChangeOutboxRepository } from './infra/postgres/postgres-route-change-outbox-repository.js';
import { PostgresUpstreamAdminRepository } from './infra/postgres/postgres-upstream-admin-repository.js';
import { ValkeyRouteProjectionRepository } from './infra/valkey/valkey-route-projection-repository.js';
import { RouteOutboxWorker } from './worker/route-outbox-worker.js';

const routeKeyPrefix = 'UPSTREAM:';
const routeChannel = 'ROUTE_OPERATIONS';

async function main() {
  const config = loadConfig();
  const pool = createPostgresPool(config.POSTGRES_CONNECTION_STRING);
  const valkey = new Redis(config.VALKEY_URL);

  const adminRepository = new PostgresUpstreamAdminRepository(pool);
  const outboxRepository = new PostgresRouteChangeOutboxRepository(pool);
  const projectionRepository = new ValkeyRouteProjectionRepository(
    valkey,
    routeKeyPrefix,
    routeChannel,
  );
  const worker = new RouteOutboxWorker(
    outboxRepository,
    projectionRepository,
    config.OUTBOX_POLL_INTERVAL_MS,
    config.OUTBOX_BATCH_SIZE,
  );
  const service = new UpstreamAdminService(adminRepository);
  const app = createApp(service);
  const snapshots = await service.listDocuments();
  for (const snapshot of snapshots) {
    await projectionRepository.setSnapshot(snapshot.serviceName, toGatewaySnapshotJson(snapshot));
  }
  const server = app.listen(config.PORT, () => {
    console.log(`console backend listening on ${config.PORT}`);
  });

  worker.start();

  const shutdown = async () => {
    await worker.stop();
    await new Promise<void>((resolve, reject) => {
      server.close((error) => {
        if (error) {
          reject(error);
          return;
        }
        resolve();
      });
    });
    await valkey.quit();
    await pool.end();
  };

  process.on('SIGINT', () => {
    void shutdown().finally(() => process.exit(0));
  });
  process.on('SIGTERM', () => {
    void shutdown().finally(() => process.exit(0));
  });
}

void main();
