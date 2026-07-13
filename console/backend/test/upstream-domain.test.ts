import { describe, expect, it } from 'vitest';
import { previewMatch, supportedJwtAlgorithms, toGatewaySnapshot, upstreamServiceSchema } from '../src/domain/upstream.js';

const service = upstreamServiceSchema.parse({
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
          rateLimitCount: 0,
        },
      ],
    },
  ],
});

describe('upstream domain mapping', () => {
  it('maps camelCase input to gateway-controller snapshot format', () => {
    expect(toGatewaySnapshot(service)).toEqual({
      service_name: 'member-api',
      authorization: {
        algorithm: 'RS256',
        key_data: 'key',
        user_key: 'user_id',
      },
      resources: [
        {
          domain: 'users',
          host: 'member.internal:8080',
          paths: [
            {
              path: '/{id}',
              method: 'GET',
              request_timeout: 3000,
              response_timeout: 5000,
              check_authorization: true,
              cache_timeout: 0,
              rate_limit_count: 0,
            },
          ],
        },
      ],
    });
  });

  it('matches gateway path with parameterized route', () => {
    expect(previewMatch(service, '/member-api/v1/users/123', 'GET')).toMatchObject({
      matched: true,
      host: 'member.internal:8080',
      upstreamPath: '/{id}',
      method: 'GET',
    });
  });

  it('supports the legacy version-first gateway path', () => {
    expect(previewMatch(service, '/v1/member-api/users/123', 'GET')).toMatchObject({
      matched: true,
      host: 'member.internal:8080',
      upstreamPath: '/{id}',
      method: 'GET',
    });
  });

  it('accepts the full Go-supported jwt algorithm set and uppercase custom methods', () => {
    expect(supportedJwtAlgorithms).toEqual(['RS256', 'RS512', 'ES256', 'ES512', 'HS256', 'HS512']);

    const parsed = upstreamServiceSchema.parse({
      serviceName: 'member-api',
      authorization: {
        algorithm: 'HS512',
        keyData: 'key',
        userKey: 'user_id',
      },
      resources: [
        {
          domain: 'users',
          host: 'member.internal:8080',
          paths: [
            {
              path: '/bulk',
              method: 'options',
              requestTimeout: 1000,
              responseTimeout: 1000,
              checkAuthorization: false,
              cacheTimeout: 0,
              rateLimitCount: 0,
            },
          ],
        },
      ],
    });

    expect(parsed.authorization?.algorithm).toBe('HS512');
    expect(parsed.resources[0].paths[0].method).toBe('OPTIONS');
  });
});
