import { describe, expect, it, vi } from 'vitest';
import { ValkeyRouteProjectionRepository } from '../src/infra/valkey/valkey-route-projection-repository.js';

describe('ValkeyRouteProjectionRepository', () => {
  it('sets the projection and publishes add/update events', async () => {
    const evalCommand = vi.fn(async () => 1);
    const publish = vi.fn(async () => 1);
    const repository = new ValkeyRouteProjectionRepository({ eval: evalCommand, publish } as never, 'UPSTREAM:', 'ROUTE_OPERATIONS');

    await repository.applyRouteChange({
      id: 1,
      serviceName: 'member-api',
      eventType: 'ROUTE_MESSAGE_ADD',
      snapshotJson: '{"service_name":"member-api","version":1}',
      serviceVersion: 1,
      attempts: 0,
    });

    expect(evalCommand).toHaveBeenCalled();
    expect(publish).toHaveBeenCalledWith(
      'ROUTE_OPERATIONS',
      JSON.stringify({ method: 'ROUTE_MESSAGE_ADD', service: 'member-api' }),
    );
  });

  it('deletes the projection for delete events', async () => {
    const evalCommand = vi.fn(async () => 1);
    const publish = vi.fn(async () => 1);
    const repository = new ValkeyRouteProjectionRepository({ eval: evalCommand, publish } as never, 'UPSTREAM:', 'ROUTE_OPERATIONS');

    await repository.applyRouteChange({
      id: 2,
      serviceName: 'member-api',
      eventType: 'ROUTE_MESSAGE_DELETE',
      snapshotJson: null,
      serviceVersion: 2,
      attempts: 0,
    });

    expect(evalCommand).toHaveBeenCalled();
    expect(publish).toHaveBeenCalledTimes(1);
  });

  it('sets snapshots directly during bootstrap without publishing', async () => {
    const evalCommand = vi.fn(async () => 1);
    const publish = vi.fn(async () => 1);
    const repository = new ValkeyRouteProjectionRepository({ eval: evalCommand, publish } as never, 'UPSTREAM:', 'ROUTE_OPERATIONS');

    await repository.setSnapshot('member-api', '{"service_name":"member-api","version":1}', 1);

    expect(evalCommand).toHaveBeenCalled();
    expect(publish).not.toHaveBeenCalled();
  });
});
