import type { RouteChangeOutboxRepository } from '../application/ports.js';
import { ValkeyRouteProjectionRepository } from '../infra/valkey/valkey-route-projection-repository.js';

export class RouteOutboxWorker {
  private timer: NodeJS.Timeout | null = null;
  private running = false;
  private inFlightFlush: Promise<void> | null = null;

  constructor(
    private readonly repository: RouteChangeOutboxRepository,
    private readonly projectionRepository: ValkeyRouteProjectionRepository,
    private readonly pollIntervalMs: number,
    private readonly batchSize: number,
  ) {}

  start() {
    if (this.timer) {
      return;
    }

    this.timer = setInterval(() => {
      void this.flush();
    }, this.pollIntervalMs);

    void this.flush();
  }

  async stop() {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }
    if (this.inFlightFlush) {
      await this.inFlightFlush;
    }
  }

  private async flush() {
    if (this.running) {
      return;
    }

    this.inFlightFlush = (async () => {
      this.running = true;
      try {
        const records = await this.repository.listPending(this.batchSize);
        for (const record of records) {
          try {
            await this.projectionRepository.applyRouteChange(record);
            await this.repository.markPublished(record.id);
          } catch (error) {
            const message = error instanceof Error ? error.message : 'unknown publish error';
            await this.repository.markFailed(record.id, message);
          }
        }
      } finally {
        this.running = false;
        this.inFlightFlush = null;
      }
    })();

    await this.inFlightFlush;
  }
}
