import { describe, expect, it, vi } from 'vitest';
import type { RouteChangeOutboxRepository } from '../src/application/ports.js';
import { RouteOutboxWorker } from '../src/worker/route-outbox-worker.js';

describe('RouteOutboxWorker', () => {
  it('waits for the in-flight flush during stop', async () => {
    let resume: (() => void) | undefined;
    const repository: RouteChangeOutboxRepository = {
      listPending: vi.fn(async () => {
        await new Promise<void>((resolve) => {
          resume = resolve;
        });
        return [];
      }),
      markPublished: vi.fn(async () => undefined),
      markFailed: vi.fn(async () => undefined),
    };
    const projectionRepository = {
      applyRouteChange: vi.fn(async () => undefined),
    } as never;
    const worker = new RouteOutboxWorker(repository, projectionRepository, 1000, 10);

    worker.start();
    const stopPromise = worker.stop();
    expect(repository.listPending).toHaveBeenCalledTimes(1);

    resume?.();
    await stopPromise;
  });
});
