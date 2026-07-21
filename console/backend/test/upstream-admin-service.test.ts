import { describe, expect, it, vi } from 'vitest';
import { UpstreamAdminService } from '../src/application/upstream-admin-service.js';
import type { UpstreamAdminRepository } from '../src/application/ports.js';
import type { UpstreamServiceDocument } from '../src/domain/upstream.js';

const sampleService: UpstreamServiceDocument = {
  serviceName: 'member-api',
  authorization: {
    algorithm: 'RS256',
    keyData: 'encoded-key',
    userKey: 'user_id',
  },
  resources: [
    {
      domain: 'users',
      description: '',
      host: 'member.internal:8080',
      paths: [
        {
          path: '/{id}',
          description: '',
          method: 'GET',
          requestTimeout: 3000,
          responseTimeout: 5000,
          checkAuthorization: true,
          cacheTimeout: 0,
          rateLimitCount: 0,
        },
      ],
    },
  ],
};

function createRepository(): UpstreamAdminRepository {
  return {
    list: vi.fn(async () => []),
    listDocuments: vi.fn(async () => []),
    get: vi.fn(async () => null),
    create: vi.fn(async () => undefined),
    update: vi.fn(async () => true),
    delete: vi.fn(async () => true),
    republish: vi.fn(async () => true),
  };
}

describe('UpstreamAdminService', () => {
  it('creates a validated service document', async () => {
    const repository = createRepository();
    const service = new UpstreamAdminService(repository);

    await expect(service.create(sampleService)).resolves.toEqual(sampleService);
    expect(repository.create).toHaveBeenCalledTimes(1);
  });

  it('rejects duplicate paths because Go routing is path-only', async () => {
    const repository = createRepository();
    const service = new UpstreamAdminService(repository);
    const invalid = {
      ...sampleService,
      resources: [
        {
          ...sampleService.resources[0],
          paths: [sampleService.resources[0].paths[0], sampleService.resources[0].paths[0]],
        },
      ],
    };

    await expect(service.create(invalid)).rejects.toThrow(/path must be unique within a resource/i);
  });

  it('maps postgres unique violations to conflict errors', async () => {
    const repository = createRepository();
    repository.create = vi.fn(async () => {
      const error = new Error('duplicate');
      (error as Error & { code?: string }).code = '23505';
      throw error;
    });
    const service = new UpstreamAdminService(repository);

    await expect(service.create(sampleService)).rejects.toThrow(/already exists/i);
  });

  it('upserts resource by previousDomain and updates only when changed', async () => {
    const repository = createRepository();
    repository.get = vi.fn(async () => sampleService);
    const service = new UpstreamAdminService(repository);

    const changedResource = {
      ...sampleService.resources[0],
      host: 'member-v2.internal:8080',
    };

    await expect(service.upsertResource('member-api', {
      previousDomain: 'users',
      resource: changedResource,
    })).resolves.toEqual({
      ...sampleService,
      resources: [changedResource],
    });
    expect(repository.update).toHaveBeenCalledTimes(1);
  });

  it('skips update when upsert payload is unchanged', async () => {
    const repository = createRepository();
    repository.get = vi.fn(async () => sampleService);
    const service = new UpstreamAdminService(repository);

    await expect(service.upsertResource('member-api', {
      previousDomain: 'users',
      resource: sampleService.resources[0],
    })).resolves.toEqual(sampleService);
    expect(repository.update).not.toHaveBeenCalled();
  });
});
