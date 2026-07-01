import { describe, expect, it, vi } from 'vitest';
import { ValkeyRouteProjectionRepository } from '../src/infra/valkey/valkey-route-projection-repository.js';

describe('ValkeyRouteProjectionRepository', () => {
  it('sets the projection and publishes add/update events', async () => {
    const set = vi.fn(async () => 'OK');
    const publish = vi.fn(async () => 1);
    const repository = new ValkeyRouteProjectionRepository({ set, publish, del: vi.fn() } as never, 'UPSTREAM:', 'ROUTE_OPERATIONS');

    await repository.applyRouteChange({
      id: 1,
      serviceName: 'member-api',
      eventType: 'ROUTE_MESSAGE_ADD',
      snapshotJson: '{"service_name":"member-api"}',
      attempts: 0,
    });

    expect(set).toHaveBeenCalledWith('UPSTREAM:member-api', '{"service_name":"member-api"}');
    expect(publish).toHaveBeenCalledWith(
      'ROUTE_OPERATIONS',
      JSON.stringify({ method: 'ROUTE_MESSAGE_ADD', service: 'member-api' }),
    );
  });

  it('deletes the projection for delete events', async () => {
    const del = vi.fn(async () => 1);
    const publish = vi.fn(async () => 1);
    const repository = new ValkeyRouteProjectionRepository({ set: vi.fn(), publish, del } as never, 'UPSTREAM:', 'ROUTE_OPERATIONS');

    await repository.applyRouteChange({
      id: 2,
      serviceName: 'member-api',
      eventType: 'ROUTE_MESSAGE_DELETE',
      snapshotJson: null,
      attempts: 0,
    });

    expect(del).toHaveBeenCalledWith('UPSTREAM:member-api');
    expect(publish).toHaveBeenCalledTimes(1);
  });

  it('sets snapshots directly during bootstrap without publishing', async () => {
    const set = vi.fn(async () => 'OK');
    const publish = vi.fn(async () => 1);
    const repository = new ValkeyRouteProjectionRepository({ set, publish, del: vi.fn() } as never, 'UPSTREAM:', 'ROUTE_OPERATIONS');

    await repository.setSnapshot('member-api', '{"service_name":"member-api"}');

    expect(set).toHaveBeenCalledWith('UPSTREAM:member-api', '{"service_name":"member-api"}');
    expect(publish).not.toHaveBeenCalled();
  });
});
