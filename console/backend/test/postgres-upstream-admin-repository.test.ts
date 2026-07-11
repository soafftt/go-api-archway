import { describe, expect, it, vi } from 'vitest';
import { PostgresUpstreamAdminRepository } from '../src/infra/postgres/postgres-upstream-admin-repository.js';

describe('PostgresUpstreamAdminRepository', () => {
  it('writes service/resources/paths and outbox records in a single transaction', async () => {
    const query = vi
      .fn()
      .mockResolvedValueOnce({ rows: [] })
      .mockResolvedValueOnce({ rows: [{ id: 11 }] })
      .mockResolvedValueOnce({ rows: [{ id: 21 }] })
      .mockResolvedValueOnce({ rows: [] })
      .mockResolvedValueOnce({ rows: [] })
      .mockResolvedValueOnce({ rows: [] });
    const release = vi.fn();
    const client = { query, release };
    const pool = {
      connect: vi.fn(async () => client),
    } as never;
    const repository = new PostgresUpstreamAdminRepository(pool);

    await repository.create(
      {
        serviceName: 'member-api',
        authorization: {
          algorithm: 'RS256',
          keyData: 'key',
          userKey: 'user_id',
        },
        resources: [
          {
            domain: 'users',
            host: 'member.internal:8080',
            paths: [
              {
                path: '/{id}',
                method: 'GET',
                requestTimeout: 3000,
                responseTimeout: 5000,
                checkAuthorization: true,
                cacheTimeout: 0,
              },
            ],
          },
        ],
      },
      '{"service_name":"member-api"}',
    );

    expect(query).toHaveBeenNthCalledWith(1, 'BEGIN');
    expect(query).toHaveBeenLastCalledWith('COMMIT');
    expect(query).toHaveBeenCalledWith(expect.stringContaining('INSERT INTO upstream_services'), ['member-api', 'RS256', 'key', 'user_id']);
    expect(query).toHaveBeenCalledWith(expect.stringContaining('INSERT INTO upstream_resources'), [11, 'users', 'member.internal:8080', 0]);
    expect(query).toHaveBeenCalledWith(expect.stringContaining('INSERT INTO route_change_outbox'), [
      'member-api',
      'ROUTE_MESSAGE_ADD',
      '{"service_name":"member-api"}',
    ]);
    expect(release).toHaveBeenCalledTimes(1);
  });

  it('lists stored documents by resolving service aggregates', async () => {
    const query = vi
      .fn()
      .mockResolvedValueOnce({ rows: [{ service_name: 'member-api' }] })
      .mockResolvedValueOnce({ rows: [{ id: 1, service_name: 'member-api', auth_algorithm: 'RS256', auth_key_data: 'key', auth_user_key: 'user_id', updated_at: new Date() }] })
      .mockResolvedValueOnce({ rows: [{ id: 10, domain: 'users', host: 'member.internal:8080', sort_order: 0 }] })
      .mockResolvedValueOnce({ rows: [{ resource_id: 10, path: '/{id}', method: 'GET', request_timeout: 3000, response_timeout: 5000, check_authorization: true, cache_timeout: 0, sort_order: 0 }] });
    const pool = { connect: vi.fn(), query } as never;
    const repository = new PostgresUpstreamAdminRepository(pool);

    const documents = await repository.listDocuments();

    expect(documents).toHaveLength(1);
    expect(documents[0].serviceName).toBe('member-api');
    expect(documents[0].resources[0].paths[0].method).toBe('GET');
  });
});
