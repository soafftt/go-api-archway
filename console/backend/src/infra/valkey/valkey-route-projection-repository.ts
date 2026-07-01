import type { Redis } from 'ioredis';
import type { RouteChangeRecord } from '../../domain/upstream.js';

export class ValkeyRouteProjectionRepository {
  constructor(
    private readonly client: Redis,
    private readonly keyPrefix: string,
    private readonly channel: string,
  ) {}

  async applyRouteChange(change: RouteChangeRecord): Promise<void> {
    const key = `${this.keyPrefix}${change.serviceName}`;

    if (change.eventType === 'ROUTE_MESSAGE_DELETE') {
      await this.client.del(key);
    } else {
      if (!change.snapshotJson) {
        throw new Error(`snapshot_json is required for ${change.eventType}`);
      }
      await this.client.set(key, change.snapshotJson);
    }

    await this.client.publish(
      this.channel,
      JSON.stringify({
        method: change.eventType,
        service: change.serviceName,
      }),
    );
  }

  async setSnapshot(serviceName: string, snapshotJson: string): Promise<void> {
    await this.client.set(`${this.keyPrefix}${serviceName}`, snapshotJson);
  }
}
