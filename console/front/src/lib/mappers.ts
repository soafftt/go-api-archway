import type { GatewaySample, UpstreamResourceDraft, UpstreamServiceDraft } from '../types';

type GatewayPreview = {
  service_name: string;
  authorization?: {
    algorithm: string;
    key_data: string;
    user_key: string;
  };
  resources: Array<{
    domain: string;
    host: string;
    paths: Array<{
      path: string;
      method: string;
      request_timeout: number;
      response_timeout: number;
      check_authorization: boolean;
      cache_timeout: number;
      rate_limit_count: number;
    }>;
  }>;
};

export function toGatewayPreview(draft: UpstreamServiceDraft): GatewayPreview {
  return {
    service_name: draft.serviceName,
    authorization: draft.authorization
      ? {
          algorithm: draft.authorization.algorithm,
          key_data: draft.authorization.keyData,
          user_key: draft.authorization.userKey,
        }
      : undefined,
    resources: draft.resources.map((resource) => ({
      domain: resource.domain,
      host: resource.host,
      paths: resource.paths.map((path) => ({
        path: path.path,
        method: path.method,
        request_timeout: path.requestTimeout,
        response_timeout: path.responseTimeout,
        check_authorization: path.checkAuthorization,
        cache_timeout: path.cacheTimeout,
        rate_limit_count: path.useRateLimit ? path.rateLimitCount : 0,
      })),
    })),
  };
}

export function normalizeDraftFromApi(draft: UpstreamServiceDraft): UpstreamServiceDraft {
  return {
    ...draft,
    resources: draft.resources.map((resource) => ({
      ...resource,
      description: resource.description ?? '',
      paths: resource.paths.map((path) => {
        const rateLimitCount = Number(path.rateLimitCount ?? 0);
        return {
          ...path,
          description: path.description ?? '',
          rateLimitCount,
          useRateLimit: rateLimitCount > 0,
        };
      }),
    })),
  };
}

export function buildGatewaySamples(service: UpstreamServiceDraft): GatewaySample[] {
  return service.resources
    .flatMap((resource) => buildResourceSamples(service.serviceName, resource))
    .sort((a, b) => {
      const byMethod = getMethodCrudOrder(a.method) - getMethodCrudOrder(b.method);
      if (byMethod !== 0) {
        return byMethod;
      }
      return a.gatewayPath.localeCompare(b.gatewayPath);
    });
}

function buildResourceSamples(serviceName: string, resource: UpstreamResourceDraft): GatewaySample[] {
  return resource.paths.map((path) => {
    const trimmedPath = path.path.replace(/^\/+/, '');
    const prefix = `/${serviceName || '{service}'}/v1`;
    const gatewayPath = resource.domain.trim() === ''
      ? (trimmedPath === '' ? `${prefix}/{domain-required}` : `${prefix}/{domain-required}/${trimmedPath}`.replace(/\/+$/, ''))
      : `${prefix}/${resource.domain}/${trimmedPath}`.replace(/\/+$/, '');

    return {
      method: path.method,
      gatewayPath,
      display: `${path.method} ${gatewayPath}`,
    };
  });
}

function getMethodCrudOrder(method: string): number {
  const normalized = method.toUpperCase();
  if (normalized === 'POST') return 0;
  if (normalized === 'GET') return 1;
  if (normalized === 'PUT' || normalized === 'PATCH') return 2;
  if (normalized === 'DELETE') return 3;
  if (normalized === 'HEAD') return 4;
  if (normalized === 'OPTIONS') return 5;

  return 99;
}
