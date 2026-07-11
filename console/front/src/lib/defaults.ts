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
    method: 'GET',
    requestTimeout: 3000,
    responseTimeout: 5000,
    checkAuthorization: false,
    cacheTimeout: 0,
  };
}

export function createResourceDraft(): UpstreamResourceDraft {
  return {
    domain: '',
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
