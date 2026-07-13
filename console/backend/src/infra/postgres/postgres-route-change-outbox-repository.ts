import type { Pool } from 'pg';
import type { RouteChangeOutboxRepository } from '../../application/ports.js';
import type { RouteChangeRecord } from '../../domain/upstream.js';

const PROCESSING_LEASE_SECONDS = 30;

export class PostgresRouteChangeOutboxRepository implements RouteChangeOutboxRepository {
  constructor(private readonly pool: Pool) {}

  async listPending(limit: number): Promise<RouteChangeRecord[]> {
    const result = await this.pool.query<RouteChangeRecord>(
      `
        WITH claimed AS (
          SELECT id
          FROM route_change_outbox
          WHERE status = 'pending'
             OR (status = 'processing' AND updated_at < CURRENT_TIMESTAMP - ($2 * INTERVAL '1 second'))
          ORDER BY id ASC
          LIMIT $1
          FOR UPDATE SKIP LOCKED
        )
        UPDATE route_change_outbox o
        SET status = 'processing', last_error = NULL
        FROM claimed
        WHERE o.id = claimed.id
        RETURNING o.id, o.service_name AS "serviceName", o.event_type AS "eventType", o.snapshot_json AS "snapshotJson", o.service_version AS "serviceVersion", o.attempts
      `,
      [limit, PROCESSING_LEASE_SECONDS],
    );

    return result.rows;
  }

  async markPublished(id: number): Promise<void> {
    await this.pool.query(
      `
        UPDATE route_change_outbox
        SET status = 'published', last_error = NULL, published_at = CURRENT_TIMESTAMP
        WHERE id = $1
      `,
      [id],
    );
  }

  async markFailed(id: number, errorMessage: string): Promise<void> {
    await this.pool.query(
      `
        UPDATE route_change_outbox
        SET
          attempts = attempts + 1,
          last_error = $1,
          status = CASE WHEN attempts + 1 >= 5 THEN 'failed' ELSE 'pending' END
        WHERE id = $2
      `,
      [errorMessage.slice(0, 1000), id],
    );
  }
}
