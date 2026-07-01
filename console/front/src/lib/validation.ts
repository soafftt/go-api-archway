import type { UpstreamServiceDraft } from '../types';

const serviceNamePattern = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
const subDomainPattern = /^[a-z0-9.-]+$/;
const hostPattern = /^[A-Za-z0-9.-]+(?::\d+)?$/;
const pathSegmentPattern = /^(?:[A-Za-z0-9._~-]+|\{[A-Za-z][A-Za-z0-9_]*\})$/;

export function validateDraft(draft: UpstreamServiceDraft): string[] {
  const errors: string[] = [];

  if (!serviceNamePattern.test(draft.serviceName)) {
    errors.push('serviceName 은 lowercase kebab-case 형식이어야 합니다.');
  }

  const seenSubDomains = new Set<string>();
  let requiresAuthorization = false;

  draft.resources.forEach((resource, resourceIndex) => {
    if (resource.subDomain !== '' && !subDomainPattern.test(resource.subDomain)) {
      errors.push(`resource #${resourceIndex + 1}: subDomain 은 소문자/숫자/점/하이픈만 허용됩니다.`);
    }

    if (seenSubDomains.has(resource.subDomain)) {
      errors.push(`resource #${resourceIndex + 1}: subDomain 은 service 내에서 중복될 수 없습니다.`);
    }
    seenSubDomains.add(resource.subDomain);

    if (!hostPattern.test(resource.host)) {
      errors.push(`resource #${resourceIndex + 1}: host 는 host[:port] 형식이어야 합니다.`);
    }

    const pathKeys = new Set<string>();
    resource.paths.forEach((path, pathIndex) => {
      if (!isValidPath(path.path)) {
        errors.push(`resource #${resourceIndex + 1} / path #${pathIndex + 1}: path 형식이 올바르지 않습니다.`);
      }
      if (path.requestTimeout <= 0 || path.responseTimeout <= 0) {
        errors.push(`resource #${resourceIndex + 1} / path #${pathIndex + 1}: timeout 은 0보다 커야 합니다.`);
      }
      if (path.cacheTimeout < 0) {
        errors.push(`resource #${resourceIndex + 1} / path #${pathIndex + 1}: cache timeout 은 0 이상이어야 합니다.`);
      }

      const key = path.path;
      if (pathKeys.has(key)) {
        errors.push(`resource #${resourceIndex + 1}: path 는 resource 내에서 중복될 수 없습니다.`);
      }
      pathKeys.add(key);

      if (path.checkAuthorization) {
        requiresAuthorization = true;
      }
    });
  });

  if (requiresAuthorization) {
    if (!draft.authorization?.keyData || !draft.authorization.userKey) {
      errors.push('checkAuthorization 경로가 있으면 authorization.keyData 와 authorization.userKey 가 필요합니다.');
    }
  }

  return errors;
}

function isValidPath(value: string) {
  if (!value.startsWith('/') || value.includes('?') || value.includes('#')) {
    return false;
  }

  return value
    .split('/')
    .filter(Boolean)
    .every((segment) => pathSegmentPattern.test(segment));
}
