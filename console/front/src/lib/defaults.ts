import type { AuthorizationDraft, UpstreamPathDraft, UpstreamResourceDraft, UpstreamServiceDraft } from '../types';

export function createAuthorizationDraft(): AuthorizationDraft {
  return {
    algorithm: 'RS256',
    keyData: '',
    userKey: 'user_id',
  };
}

export function createPathDraft(): UpstreamPathDraft {
  return {
    path: '/',
    description: '',
    method: 'GET',
    requestTimeout: 3000,
    responseTimeout: 5000,
    checkAuthorization: false,
    cacheTimeout: 0,
    useRateLimit: false,
    rateLimitCount: 0,
  };
}

export function createResourceDraft(): UpstreamResourceDraft {
  return {
    domain: '',
    description: '',
    host: '',
    paths: [createPathDraft()],
  };
}

export function createServiceDraft(): UpstreamServiceDraft {
  return {
    serviceName: '',
    resources: [createResourceDraft()],
  };
}
