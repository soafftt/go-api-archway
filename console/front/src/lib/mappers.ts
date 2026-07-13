import type { UpstreamResourceDraft, UpstreamServiceDraft } from '../types';

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
      paths: resource.paths.map((path) => {
        const rateLimitCount = Number(path.rateLimitCount ?? 0);
        return {
          ...path,
          rateLimitCount,
          useRateLimit: rateLimitCount > 0,
        };
      }),
    })),
  };
}

export function buildGatewaySamples(service: UpstreamServiceDraft): string[] {
  return service.resources.flatMap((resource) => buildResourceSamples(service.serviceName, resource));
}

function buildResourceSamples(serviceName: string, resource: UpstreamResourceDraft): string[] {
  return resource.paths.map((path) => {
    const trimmedPath = path.path.replace(/^\/+/, '');
    const prefix = `/${serviceName || '{service}'}/v1`;
    if (resource.domain.trim() === '') {
      if (trimmedPath === '') {
        return `${prefix}/{domain-required}`;
      }
      return `${prefix}/{domain-required}/${trimmedPath}`.replace(/\/+$/, '');
    }

    return `${prefix}/${resource.domain}/${trimmedPath}`.replace(/\/+$/, '');
  });
}
