import { z } from 'zod';

export const routeChangeTypes = ['ROUTE_MESSAGE_ADD', 'ROUTE_MESSAGE_UPDATE', 'ROUTE_MESSAGE_DELETE'] as const;
export const supportedJwtAlgorithms = ['RS256', 'RS512', 'ES256', 'ES512', 'HS256', 'HS512'] as const;

export type RouteChangeType = (typeof routeChangeTypes)[number];

const authorizationSchema = z.object({
  algorithm: z.enum(supportedJwtAlgorithms),
  keyData: z.string().min(1, 'authorization.keyData is required'),
  userKey: z.string().min(1, 'authorization.userKey is required'),
});

const pathVariablePattern = /^\{[A-Za-z][A-Za-z0-9_]*\}$/;

function isValidRoutePath(value: string): boolean {
  if (!value.startsWith('/')) {
    return false;
  }
  if (value.includes('?') || value.includes('#')) {
    return false;
  }

  return value
    .split('/')
    .filter(Boolean)
    .every((segment) => pathVariablePattern.test(segment) || /^[A-Za-z0-9._~-]+$/.test(segment));
}

const upstreamPathSchema = z.object({
  path: z.string().min(1).refine(isValidRoutePath, 'path must start with "/" and only contain literal or {variable} segments'),
  description: z.string().nullable().optional().transform((value) => normalizeDescription(value)),
  method: z.string().min(1).transform((value) => value.toUpperCase()).refine((value) => /^[A-Z]+$/.test(value), 'method must be an uppercase token'),
  requestTimeout: z.coerce.number().int().positive(),
  responseTimeout: z.coerce.number().int().positive(),
  checkAuthorization: z.boolean(),
  cacheTimeout: z.coerce.number().int().min(0),
  rateLimitCount: z.coerce.number().int().min(0),
});

export const upstreamResourceSchema = z.object({
  domain: z
    .string()
    .refine((value) => value === '' || /^[a-z0-9.-]+$/.test(value), 'domain must be empty or lowercase letters, numbers, dots, hyphens'),
  description: z.string().nullable().optional().transform((value) => normalizeDescription(value)),
  host: z.string().refine((value) => /^[A-Za-z0-9.-]+(?::\d+)?$/.test(value), 'host must be host[:port] without scheme'),
  paths: z.array(upstreamPathSchema).min(1, 'at least one path is required'),
});

export const upstreamServiceSchema = z
  .object({
    serviceName: z.string().regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/, 'serviceName must be lowercase kebab-case'),
    authorization: authorizationSchema.optional(),
    resources: z.array(upstreamResourceSchema).min(1, 'at least one resource is required'),
  })
  .superRefine((service, ctx) => {
    const domains = new Set<string>();
    let requiresAuthorization = false;

    service.resources.forEach((resource, resourceIndex) => {
      if (domains.has(resource.domain)) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ['resources', resourceIndex, 'domain'],
          message: 'domain must be unique within a service',
        });
      }
      domains.add(resource.domain);

      const pathKeys = new Set<string>();
      resource.paths.forEach((path, pathIndex) => {
        const pathKey = path.path;
        if (pathKeys.has(pathKey)) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            path: ['resources', resourceIndex, 'paths', pathIndex, 'path'],
            message: 'path must be unique within a resource because the Go controller routes by path only',
          });
        }
        pathKeys.add(pathKey);

        if (path.checkAuthorization) {
          requiresAuthorization = true;
        }
      });
    });

    if (requiresAuthorization && !service.authorization) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['authorization'],
        message: 'authorization is required when any path enables checkAuthorization',
      });
    }
  });

export type UpstreamAuthorization = z.infer<typeof authorizationSchema>;
export type UpstreamPath = z.infer<typeof upstreamPathSchema>;
export type UpstreamResource = z.infer<typeof upstreamResourceSchema>;
export type UpstreamServiceInput = z.infer<typeof upstreamServiceSchema>;
export type UpstreamServiceDocument = UpstreamServiceInput & {
  version?: number;
};

export type UpstreamServiceSummary = {
  serviceName: string;
  resourceCount: number;
  pathCount: number;
  updatedAt?: string;
};

export type RouteChangeRecord = {
  id: number;
  serviceName: string;
  eventType: RouteChangeType;
  snapshotJson: string | null;
  serviceVersion: number;
  attempts: number;
};

export type PreviewMatchResult = {
  version: string;
  serviceName: string;
  domain: string;
  matched: boolean;
  host?: string;
  upstreamPath?: string;
  method?: string;
};

export function toGatewaySnapshot(service: UpstreamServiceDocument, version?: number) {
  const resolvedVersion = version ?? service.version;
  return {
    service_name: service.serviceName,
    ...(typeof resolvedVersion === 'number' ? { version: resolvedVersion } : {}),
    authorization: service.authorization
      ? {
          algorithm: service.authorization.algorithm,
          key_data: service.authorization.keyData,
          user_key: service.authorization.userKey,
        }
      : undefined,
    resources: service.resources.map((resource) => ({
      domain: resource.domain,
      host: resource.host,
      paths: resource.paths.map((path) => ({
        path: path.path,
        method: path.method,
        request_timeout: path.requestTimeout,
        response_timeout: path.responseTimeout,
        check_authorization: path.checkAuthorization,
        cache_timeout: path.cacheTimeout,
        rate_limit_count: path.rateLimitCount,
      })),
    })),
  };
}

export function toGatewaySnapshotJson(service: UpstreamServiceDocument, version?: number): string {
  return JSON.stringify(toGatewaySnapshot(service, version));
}

export function summarizeService(service: UpstreamServiceDocument, updatedAt?: string): UpstreamServiceSummary {
  return {
    serviceName: service.serviceName,
    resourceCount: service.resources.length,
    pathCount: service.resources.reduce((total, resource) => total + resource.paths.length, 0),
    updatedAt,
  };
}

export function previewMatch(service: UpstreamServiceDocument, gatewayPath: string, method: string): PreviewMatchResult {
  const normalized = parseGatewayPath(gatewayPath);
  if (!normalized.isValid || normalized.serviceName !== service.serviceName) {
    return { ...normalized, matched: false };
  }

  const directResource = service.resources.find((resource) => resource.domain === normalized.domain);
  const fallbackResource = service.resources.find((resource) => resource.domain === '');
  const resource = directResource ?? fallbackResource;
  if (!resource) {
    return { ...normalized, matched: false };
  }

  const requestPath = directResource ? normalized.remainingPath : `/${[normalized.domain, normalized.remainingPath].filter(Boolean).join('/')}`;
  const matchedPath = resource.paths.find((path) => matchesRoutePattern(path.path, requestPath));

  if (!matchedPath) {
    return { ...normalized, matched: false };
  }

  return {
    ...normalized,
    matched: true,
    host: resource.host,
    upstreamPath: matchedPath.path,
    method: matchedPath.method,
  };
}

function parseGatewayPath(input: string) {
  const path = new URL(input, 'http://localhost').pathname;
  const segments = path.split('/').filter(Boolean);
  const isLegacyPath = /^v\d+$/.test(segments[0] ?? '');
  const serviceName = isLegacyPath ? (segments[1] ?? '') : (segments[0] ?? '');
  const version = isLegacyPath ? (segments[0] ?? '') : (segments[1] ?? '');
  const domain = segments[2] ?? '';
  const rest = segments.slice(3);

  return {
    isValid: segments.length >= 4,
    version,
    serviceName,
    domain,
    remainingPath: `/${rest.join('/')}`.replace(/\/+$/, '') || '/',
  };
}

function matchesRoutePattern(pattern: string, requestPath: string): boolean {
  const requestSegments = requestPath.split('/').filter(Boolean);
  const patternSegments = pattern.split('/').filter(Boolean);

  if (requestSegments.length !== patternSegments.length) {
    return false;
  }

  for (let index = 0; index < patternSegments.length; index += 1) {
    const patternSegment = patternSegments[index];
    const requestSegment = requestSegments[index];

    if (patternSegment === requestSegment) {
      continue;
    }
    if (pathVariablePattern.test(patternSegment)) {
      continue;
    }
    return false;
  }

  return true;
}

function normalizeDescription(value: string | null | undefined): string {
  if (value === null || value === undefined) {
    return '';
  }
  return value;
}
