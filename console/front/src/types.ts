export type AuthorizationDraft = {
  algorithm: 'RS256' | 'RS512' | 'ES256' | 'ES512' | 'HS256' | 'HS512';
  keyData: string;
  userKey: string;
};

export type UpstreamPathDraft = {
  path: string;
  description: string;
  method: string;
  requestTimeout: number;
  responseTimeout: number;
  checkAuthorization: boolean;
  cacheTimeout: number;
  useRateLimit: boolean;
  rateLimitCount: number;
};

export type UpstreamResourceDraft = {
  domain: string;
  description: string;
  host: string;
  paths: UpstreamPathDraft[];
};

export type UpstreamServiceDraft = {
  serviceName: string;
  authorization?: AuthorizationDraft;
  resources: UpstreamResourceDraft[];
};

export type UpstreamServiceSummary = {
  serviceName: string;
  resourceCount: number;
  pathCount: number;
  updatedAt?: string;
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

export type GatewaySample = {
  method: string;
  gatewayPath: string;
  display: string;
};
