import { describe, expect, it, vi } from 'vitest';
import { PostgresRouteChangeOutboxRepository } from '../src/infra/postgres/postgres-route-change-outbox-repository.js';

describe('PostgresRouteChangeOutboxRepository', () => {
  it('claims pending or stale processing rows with a processing lease', async () => {
    const query = vi.fn(async () => ({
      rows: [
        {
          id: 1,
          serviceName: 'member-api',
          eventType: 'ROUTE_MESSAGE_UPDATE',
          snapshotJson: '{"service_name":"member-api"}',
          attempts: 0,
        },
      ],
    }));
    const repository = new PostgresRouteChangeOutboxRepository({ query } as never);

    const rows = await repository.listPending(10);

    expect(rows).toHaveLength(1);
    expect(query).toHaveBeenCalledWith(expect.stringContaining("status = 'processing' AND updated_at < CURRENT_TIMESTAMP"), [10, 30]);
  });

  it('marks failures with a truncated error message', async () => {
    const query = vi.fn(async () => ({ rows: [] }));
    const repository = new PostgresRouteChangeOutboxRepository({ query } as never);

    await repository.markFailed(7, 'x'.repeat(1205));

    const [, params] = query.mock.calls[0];
    expect(params[0]).toHaveLength(1000);
    expect(params[1]).toBe(7);
  });
});
