import type { UpstreamResourceDraft, UpstreamServiceDraft } from '../types';

type GatewayPreview = {
  service_name: string;
  authorization?: {
    algorithm: string;
    key_data: string;
    user_key: string;
  };
  resources: Array<{
    sub_domain: string;
    host: string;
    paths: Array<{
      path: string;
      method: string;
      request_timeout: number;
      response_timeout: number;
      check_authorization: boolean;
      cache_timeout: number;
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
      sub_domain: resource.subDomain,
      host: resource.host,
      paths: resource.paths.map((path) => ({
        path: path.path,
        method: path.method,
        request_timeout: path.requestTimeout,
        response_timeout: path.responseTimeout,
        check_authorization: path.checkAuthorization,
        cache_timeout: path.cacheTimeout,
      })),
    })),
  };
}

export function buildGatewaySamples(service: UpstreamServiceDraft): string[] {
  return service.resources.flatMap((resource) => buildResourceSamples(service.serviceName, resource));
}

function buildResourceSamples(serviceName: string, resource: UpstreamResourceDraft): string[] {
  return resource.paths.map((path) => {
    const trimmedPath = path.path.replace(/^\/+/, '');
    const prefix = `/v1/${serviceName || '{service}'}`;
    if (resource.subDomain.trim() === '') {
      if (trimmedPath === '') {
        return `${prefix}/{domain-required}`;
      }
      return `${prefix}/${trimmedPath}`.replace(/\/+$/, '');
    }

    return `${prefix}/${resource.subDomain}/${trimmedPath}`.replace(/\/+$/, '');
  });
}
