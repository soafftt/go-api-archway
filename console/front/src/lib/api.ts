import type { PreviewMatchResult, UpstreamServiceDraft, UpstreamServiceSummary } from '../types';

const BASE_URL = '/api/v1/upstream-services';

async function request<T>(input: RequestInfo, init?: RequestInit): Promise<T> {
  const response = await fetch(input, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
  });

  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as { message?: string } | null;
    throw new Error(payload?.message ?? `요청 실패: ${response.status}`);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

export function listServices(): Promise<UpstreamServiceSummary[]> {
  return request<UpstreamServiceSummary[]>(BASE_URL);
}

export function getService(serviceName: string): Promise<UpstreamServiceDraft> {
  return request<UpstreamServiceDraft>(`${BASE_URL}/${serviceName}`);
}

export function createService(payload: UpstreamServiceDraft): Promise<UpstreamServiceDraft> {
  return request<UpstreamServiceDraft>(BASE_URL, {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updateService(serviceName: string, payload: UpstreamServiceDraft): Promise<UpstreamServiceDraft> {
  return request<UpstreamServiceDraft>(`${BASE_URL}/${serviceName}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}

export function deleteService(serviceName: string): Promise<void> {
  return request<void>(`${BASE_URL}/${serviceName}`, {
    method: 'DELETE',
  });
}

export function previewMatch(gatewayPath: string): Promise<PreviewMatchResult> {
  return request<PreviewMatchResult>(`${BASE_URL}/preview-match`, {
    method: 'POST',
    body: JSON.stringify({ gatewayPath, method: 'GET' }),
  });
}
