import { useEffect, useMemo, useState } from 'react';
import { previewMatch } from '../lib/api';
import { buildGatewaySamples, toGatewayPreview } from '../lib/mappers';
import { validateDraft } from '../lib/validation';
import { createAuthorizationDraft, createPathDraft, createResourceDraft } from '../lib/defaults';
import type { AuthorizationDraft, PreviewMatchResult, UpstreamPathDraft, UpstreamResourceDraft, UpstreamServiceDraft } from '../types';

type ServiceEditorProps = {
  draft: UpstreamServiceDraft;
  baselineDraft: UpstreamServiceDraft | null;
  mode: 'create' | 'edit';
  busy: boolean;
  message: string | null;
  error: string | null;
  onChange: (next: UpstreamServiceDraft) => void;
  onSaveResource: (resource: UpstreamResourceDraft, previousDomain: string) => void;
  onDelete: () => void;
};

const methods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'] as const;
const algorithms = ['RS256', 'RS512', 'ES256', 'ES512', 'HS256', 'HS512'] as const;

export function ServiceEditor({
  draft,
  baselineDraft,
  mode,
  busy,
  message,
  error,
  onChange,
  onSaveResource,
  onDelete,
}: ServiceEditorProps) {
  const previewJson = JSON.stringify(toGatewayPreview(draft), null, 2);
  const samples = buildGatewaySamples(draft);
  const validationErrors = useMemo(() => validateDraft(draft), [draft]);
  const [gatewayPath, setGatewayPath] = useState('');
  const [selectedSampleIndex, setSelectedSampleIndex] = useState(0);
  const [selectedResourceIndex, setSelectedResourceIndex] = useState(0);
  const [selectedResourcePreviousDomain, setSelectedResourcePreviousDomain] = useState('');
  const [matchResult, setMatchResult] = useState<PreviewMatchResult | null>(null);
  const [matchError, setMatchError] = useState<string | null>(null);
  const [matching, setMatching] = useState(false);

  useEffect(() => {
    setMatchResult(null);
    setMatchError(null);
  }, [draft]);

  useEffect(() => {
    if (samples.length === 0) {
      setSelectedSampleIndex(0);
      return;
    }
    if (selectedSampleIndex > samples.length - 1) {
      setSelectedSampleIndex(0);
    }
  }, [samples.length, selectedSampleIndex]);

  useEffect(() => {
    if (draft.resources.length === 0) {
      setSelectedResourceIndex(0);
      setSelectedResourcePreviousDomain('');
      return;
    }
    if (selectedResourceIndex > draft.resources.length - 1) {
      const nextIndex = Math.max(0, draft.resources.length - 1);
      setSelectedResourceIndex(nextIndex);
      setSelectedResourcePreviousDomain(draft.resources[nextIndex]?.domain ?? '');
    }
  }, [draft.resources.length, selectedResourceIndex]);

  useEffect(() => {
    if (!draft.resources[selectedResourceIndex]) {
      return;
    }
    setSelectedResourcePreviousDomain(draft.resources[selectedResourceIndex].domain);
  }, [draft.resources, selectedResourceIndex]);

  const handlePreviewMatch = async () => {
    setMatching(true);
    setMatchError(null);
    try {
      const result = await previewMatch(gatewayPath);
      setMatchResult(result);
    } catch (previewError) {
      setMatchResult(null);
      setMatchError(previewError instanceof Error ? previewError.message : 'preview 요청에 실패했습니다.');
    } finally {
      setMatching(false);
    }
  };

  const bindSampleToPreview = () => {
    if (samples.length === 0) {
      return;
    }
    const sample = samples[selectedSampleIndex] ?? samples[0];
    setGatewayPath(sample.gatewayPath);
    setMatchResult(null);
    setMatchError(null);
  };

  const selectResource = (resourceIndex: number) => {
    setSelectedResourceIndex(resourceIndex);
    setSelectedResourcePreviousDomain(draft.resources[resourceIndex]?.domain ?? '');
  };

  const addResourceDomain = () => {
    const nextResources = [...draft.resources, createResourceDraft()];
    onChange({ ...draft, resources: nextResources });
    const nextIndex = nextResources.length - 1;
    setSelectedResourceIndex(nextIndex);
    setSelectedResourcePreviousDomain(nextResources[nextIndex].domain);
  };

  const selectedResource = draft.resources[selectedResourceIndex];
  const baselineSelectedResource = selectedResource
    ? findResourceByDomain(
      baselineDraft?.resources ?? [],
      selectedResourcePreviousDomain || selectedResource.domain,
    )
    : undefined;
  const isAddedSelectedResource = Boolean(selectedResource && !baselineSelectedResource);

  return (
    <section className="flex min-h-screen flex-1 flex-col bg-slate-950">
      <header className="sticky top-0 z-10 border-b border-slate-800 bg-slate-950/95 px-6 py-4 backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-xl font-semibold text-white">
              {mode === 'create' ? '새 routing rule' : draft.serviceName || 'routing rule'}
            </h2>
            <p className="text-sm text-slate-400">gateway-controller 의 Valkey projection 형식에 맞춰 편집합니다.</p>
          </div>

          {mode === 'edit' ? (
            <button
              type="button"
              onClick={onDelete}
              disabled={busy}
              className="rounded-md border border-rose-500/40 px-3 py-2 text-sm font-medium text-rose-200 transition hover:bg-rose-500/10 disabled:cursor-not-allowed disabled:opacity-60"
            >
              삭제
            </button>
          ) : null}
        </div>
        {message ? <p className="mt-3 text-sm text-emerald-300">{message}</p> : null}
        {error ? <p className="mt-3 text-sm text-rose-300">{error}</p> : null}
      </header>

      <div className="grid flex-1 gap-6 px-6 py-6 xl:grid-cols-[minmax(0,2fr)_minmax(320px,1fr)]">
        <div className="space-y-6">
          <Card title="Service">
            <div className="grid gap-4 md:grid-cols-2">
              <Field label="serviceName">
                <input
                  value={draft.serviceName}
                  disabled={mode === 'edit'}
                  onChange={(event) => onChange({ ...draft, serviceName: event.target.value })}
                  className="w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-white disabled:cursor-not-allowed disabled:opacity-70"
                  placeholder="member-api"
                />
              </Field>
              <div className="rounded-md border border-slate-800 bg-slate-900 px-3 py-2 text-sm text-slate-300">
                <div className="font-medium text-white">Gateway lookup segment</div>
                <div className="mt-1 text-xs text-slate-400">
                  현재 gateway-controller 규칙에서는 `/{'{serviceName}'}/v1/{'{domain}'}/...` 형태로 조회됩니다.
                </div>
              </div>
            </div>
          </Card>

          <Card title="Client validation">
            {validationErrors.length === 0 ? (
              <p className="text-sm text-emerald-300">현재 입력은 기본 클라이언트 검증을 통과했습니다.</p>
            ) : (
              <ul className="space-y-2 text-sm text-rose-200">
                {validationErrors.map((validationError) => (
                  <li key={validationError} className="rounded-md border border-rose-500/30 bg-rose-500/10 px-3 py-2">
                    {validationError}
                  </li>
                ))}
              </ul>
            )}
          </Card>

          <Card title="Authorization">
            <div className="mb-4">
              <label className="inline-flex items-center gap-2 text-sm text-slate-200">
                <input
                  type="checkbox"
                  checked={Boolean(draft.authorization)}
                  onChange={(event) =>
                    onChange({
                      ...draft,
                      authorization: event.target.checked ? createAuthorizationDraft() : undefined,
                    })
                  }
                />
                service authorization 사용
              </label>
            </div>
            {draft.authorization ? (
              <AuthorizationFields
                value={draft.authorization}
                onChange={(authorization) => onChange({ ...draft, authorization })}
              />
            ) : (
              <p className="text-sm text-slate-400">경로 중 `checkAuthorization` 이 켜진 항목이 있으면 authorization 이 필수입니다.</p>
            )}
          </Card>

          <Card
            title="Resources"
            action={
              <button
                type="button"
                onClick={addResourceDomain}
                className="rounded-md border border-slate-700 px-3 py-2 text-sm text-slate-100 transition hover:border-slate-500"
              >
                domain 추가
              </button>
            }
          >
            <div className="grid gap-4 lg:grid-cols-[240px_minmax(0,1fr)]">
              <aside className="max-h-[720px] overflow-y-auto rounded-xl border border-slate-800 bg-slate-950/60 p-3">
                <p className="mb-3 text-xs uppercase tracking-wide text-slate-400">Domain tree</p>
                <ul className="space-y-2">
                  {draft.resources.map((resource, resourceIndex) => {
                    const domainLabel = resource.domain.trim() === '' ? '(fallback)' : resource.domain;
                    const isSelected = selectedResourceIndex === resourceIndex;
                    const isAdded = !findResourceByDomain(baselineDraft?.resources ?? [], resource.domain);

                    return (
                      <li key={resourceIndex}>
                        <button
                          type="button"
                          onClick={() => selectResource(resourceIndex)}
                          className={`w-full rounded-lg border px-3 py-2 text-left transition ${
                            isSelected
                              ? 'border-sky-400 bg-sky-500/10 text-white'
                              : 'border-slate-800 bg-slate-900/40 text-slate-200 hover:border-slate-700'
                          }`}
                        >
                          <div className="truncate text-sm font-medium">{domainLabel}</div>
                          <div className="mt-1 truncate text-xs text-slate-400">
                            paths {resource.paths.length} · {resource.host || 'host 미지정'}
                          </div>
                          {isAdded ? <div className="mt-1 text-[11px] text-emerald-300">추가됨</div> : null}
                        </button>
                      </li>
                    );
                  })}
                </ul>
              </aside>

              <div className="max-h-[720px] overflow-y-auto pr-1">
                {selectedResource ? (
                  <div key={selectedResourceIndex}>
                    <ResourceEditor
                      resource={selectedResource}
                      baselineResource={baselineSelectedResource}
                      isAddedResource={isAddedSelectedResource}
                      index={selectedResourceIndex}
                      busy={busy}
                      onSave={() => onSaveResource(selectedResource, selectedResourcePreviousDomain)}
                      onChange={(nextResource) => {
                        const resources = [...draft.resources];
                        resources[selectedResourceIndex] = nextResource;
                        onChange({ ...draft, resources });
                      }}
                      onRemove={() => {
                        const resources = draft.resources.filter((_, index) => index !== selectedResourceIndex);
                        const nextResources = resources.length > 0 ? resources : [createResourceDraft()];
                        const nextSelectedIndex = Math.min(selectedResourceIndex, nextResources.length - 1);
                        setSelectedResourceIndex(Math.max(0, nextSelectedIndex));
                        setSelectedResourcePreviousDomain(nextResources[Math.max(0, nextSelectedIndex)]?.domain ?? '');
                        onChange({ ...draft, resources: nextResources });
                      }}
                    />
                  </div>
                ) : (
                  <div className="rounded-xl border border-dashed border-slate-700 p-4 text-sm text-slate-400">
                    편집할 domain 을 왼쪽 트리에서 선택하세요.
                  </div>
                )}
              </div>
            </div>
          </Card>
        </div>

        <div className="space-y-6">
          <Card
            title="Gateway path samples"
            action={
              <div className="flex items-center gap-2">
                <select
                  value={selectedSampleIndex}
                  onChange={(event) => setSelectedSampleIndex(Number(event.target.value))}
                  disabled={samples.length === 0}
                  className="w-72 rounded-md border border-slate-700 bg-slate-950 px-2 py-2 text-xs text-white disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {samples.map((sample, index) => (
                    <option key={`${sample.method}-${sample.gatewayPath}-${index}`} value={index}>
                      {sample.display}
                    </option>
                  ))}
                </select>
                <button
                  type="button"
                  onClick={bindSampleToPreview}
                  disabled={samples.length === 0}
                  className="rounded-md border border-slate-700 px-3 py-2 text-sm text-slate-100 transition hover:border-sky-400 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  Preview 테스트 바인드
                </button>
              </div>
            }
          >
            {samples.length === 0 ? (
              <p className="text-sm text-slate-400">domain/path 를 추가하면 샘플 path 가 생성됩니다.</p>
            ) : (
              <ul className="space-y-2 text-sm text-slate-300">
                {samples.map((sample, index) => (
                  <li key={`${sample.method}-${sample.gatewayPath}-${index}`} className="rounded-md bg-slate-950 px-3 py-2 font-mono text-xs text-sky-200">
                    {sample.display}
                  </li>
                ))}
              </ul>
            )}
          </Card>

          <Card title="Rewrite preview test">
            <div className="space-y-3">
              <p className="text-xs text-slate-400">이 패널은 저장된 규칙 기준으로 gateway-controller의 path-only 매칭 결과를 조회합니다.</p>
              <Field label="gateway path">
                <input
                  value={gatewayPath}
                  onChange={(event) => setGatewayPath(event.target.value)}
                  className="w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-white"
                  placeholder="/member-api/v1/users/123"
                />
              </Field>
              <button
                type="button"
                onClick={() => void handlePreviewMatch()}
                disabled={matching || gatewayPath.trim() === ''}
                className="rounded-md border border-slate-700 px-3 py-2 text-sm text-slate-100 transition hover:border-sky-400 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {matching ? '검사 중...' : 'preview-match 실행'}
              </button>
              {matchError ? <p className="text-sm text-rose-300">{matchError}</p> : null}
              {matchResult ? (
                <div className="rounded-md border border-slate-800 bg-slate-950 p-4 text-sm text-slate-200">
                  <div>matched: <span className={matchResult.matched ? 'text-emerald-300' : 'text-rose-300'}>{String(matchResult.matched)}</span></div>
                  <div className="mt-1">service: {matchResult.serviceName || '-'}</div>
                  <div className="mt-1">domain: {matchResult.domain || '-'}</div>
                  <div className="mt-1">host: {matchResult.host || '-'}</div>
                  <div className="mt-1">upstreamPath: {matchResult.upstreamPath || '-'}</div>
                  <div className="mt-1">stored method metadata: {matchResult.method || '-'}</div>
                </div>
              ) : null}
            </div>
          </Card>

          <Card title="Valkey snapshot preview">
            <pre className="max-h-[560px] overflow-auto rounded-md bg-slate-950 p-4 text-xs text-slate-200">{previewJson}</pre>
          </Card>
        </div>
      </div>
    </section>
  );
}

function AuthorizationFields({
  value,
  onChange,
}: {
  value: AuthorizationDraft;
  onChange: (next: AuthorizationDraft) => void;
}) {
  return (
    <div className="grid gap-4 md:grid-cols-3">
      <Field label="algorithm">
        <select
          value={value.algorithm}
          onChange={(event) => onChange({ ...value, algorithm: event.target.value as AuthorizationDraft['algorithm'] })}
          className="w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-white"
        >
          {algorithms.map((algorithm) => (
            <option key={algorithm} value={algorithm}>
              {algorithm}
            </option>
          ))}
        </select>
      </Field>
      <Field label="userKey">
        <input
          value={value.userKey}
          onChange={(event) => onChange({ ...value, userKey: event.target.value })}
          className="w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-white"
        />
      </Field>
      <Field label="keyData">
        <input
          value={value.keyData}
          onChange={(event) => onChange({ ...value, keyData: event.target.value })}
          className="w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-white"
          placeholder="base64 encoded JWK / shared secret"
        />
      </Field>
    </div>
  );
}

function ResourceEditor({
  resource,
  baselineResource,
  isAddedResource,
  index,
  busy,
  onSave,
  onChange,
  onRemove,
}: {
  resource: UpstreamResourceDraft;
  baselineResource?: UpstreamResourceDraft;
  isAddedResource: boolean;
  index: number;
  busy: boolean;
  onSave: () => void;
  onChange: (next: UpstreamResourceDraft) => void;
  onRemove: () => void;
}) {
  const domainChanged = isFieldChanged(resource.domain, baselineResource?.domain);
  const resourceDescriptionChanged = isFieldChanged(resource.description, baselineResource?.description);
  const hostChanged = isFieldChanged(resource.host, baselineResource?.host);
  const sortedPaths = resource.paths
    .map((path, index) => ({ path, index }))
    .sort((a, b) => {
      const byMethod = getMethodCrudOrder(a.path.method) - getMethodCrudOrder(b.path.method);
      if (byMethod !== 0) {
        return byMethod;
      }
      return a.path.path.localeCompare(b.path.path);
    });

  return (
    <div className={`rounded-xl border bg-slate-900/80 p-5 ${isAddedResource ? 'border-emerald-400/60' : 'border-slate-800'}`}>
      <div className="mb-4 flex items-center justify-between gap-2">
        <div>
          <h4 className="font-medium text-white">
            Resource #{index + 1}
            {isAddedResource ? <span className="ml-2 rounded-md bg-emerald-500/20 px-2 py-1 text-xs text-emerald-300">추가됨</span> : null}
          </h4>
          <p className="text-xs text-slate-400">domain 은 DDD context 키(예: users, orders)로 사용하며, 비어 있으면 fallback resource 로 저장됩니다.</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={onSave}
            disabled={busy}
            className="rounded-md bg-sky-500 px-3 py-2 text-sm font-semibold text-slate-950 transition hover:bg-sky-400 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {busy ? '저장 중...' : '이 domain 저장'}
          </button>
          <button
            type="button"
            onClick={onRemove}
            className="rounded-md border border-slate-700 px-3 py-2 text-sm text-slate-200 transition hover:border-rose-500 hover:text-rose-200"
          >
            제거
          </button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Field label="domain">
          <input
            value={resource.domain}
            onChange={(event) => onChange({ ...resource, domain: event.target.value })}
            className={buildInputClass(domainChanged, isAddedResource)}
            placeholder="users, orders, billing"
          />
        </Field>
        <Field label="domain description (nullable)">
          <input
            value={resource.description}
            onChange={(event) => onChange({ ...resource, description: event.target.value })}
            className={buildInputClass(resourceDescriptionChanged, isAddedResource)}
            placeholder="도메인 설명 (선택)"
          />
        </Field>
        <Field label="host">
          <input
            value={resource.host}
            onChange={(event) => onChange({ ...resource, host: event.target.value })}
            className={buildInputClass(hostChanged, isAddedResource)}
            placeholder="member.internal:8080"
          />
        </Field>
      </div>

      <div className="mt-5 max-h-[420px] overflow-auto">
        <table className="min-w-full text-left text-sm text-slate-200">
          <thead className="text-xs uppercase tracking-wide text-slate-400">
            <tr>
              <th className="pb-2">method</th>
              <th className="pb-2">path</th>
              <th className="pb-2">description</th>
              <th className="pb-2">req timeout</th>
              <th className="pb-2">res timeout</th>
              <th className="pb-2">cache</th>
              <th className="pb-2">rate limit</th>
              <th className="pb-2">auth</th>
              <th className="pb-2"></th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800">
            {sortedPaths.map(({ path, index: originalPathIndex }) => (
              <PathRow
                key={`${path.method}-${path.path}-${originalPathIndex}`}
                path={path}
                baselinePath={baselineResource?.paths[originalPathIndex]}
                isAdded={isAddedResource || !baselineResource?.paths[originalPathIndex]}
                onChange={(nextPath) => {
                  const paths = [...resource.paths];
                  paths[originalPathIndex] = nextPath;
                  onChange({ ...resource, paths });
                }}
                onRemove={() => {
                  const paths = resource.paths.filter((_, indexToRemove) => indexToRemove !== originalPathIndex);
                  onChange({ ...resource, paths: paths.length > 0 ? paths : [createPathDraft()] });
                }}
              />
            ))}
          </tbody>
        </table>
      </div>

      <div className="mt-4">
        <button
          type="button"
          onClick={() => onChange({ ...resource, paths: [...resource.paths, createPathDraft()] })}
          className="rounded-md border border-slate-700 px-3 py-2 text-sm text-slate-100 transition hover:border-slate-500"
        >
          path 추가
        </button>
      </div>
    </div>
  );
}

function PathRow({
  path,
  baselinePath,
  isAdded,
  onChange,
  onRemove,
}: {
  path: UpstreamPathDraft;
  baselinePath?: UpstreamPathDraft;
  isAdded: boolean;
  onChange: (next: UpstreamPathDraft) => void;
  onRemove: () => void;
}) {
  const pathChanged = isFieldChanged(path.path, baselinePath?.path);
  const descriptionChanged = isFieldChanged(path.description, baselinePath?.description);

  return (
    <tr className={isAdded ? 'bg-emerald-500/15' : undefined}>
      <td className="py-3 pr-2">
        <select
          value={path.method}
          onChange={(event) => onChange({ ...path, method: event.target.value })}
          className={buildInputClass(isFieldChanged(path.method, baselinePath?.method), isAdded, 'w-full px-2 py-2 text-xs')}
        >
          {methods.map((method) => (
            <option key={method} value={method}>
              {method}
            </option>
          ))}
        </select>
      </td>
      <td className="py-3 pr-2">
        <input
          value={path.path}
          onChange={(event) => onChange({ ...path, path: event.target.value })}
          className={buildInputClass(pathChanged, isAdded, 'w-full px-2 py-2 text-xs')}
          placeholder="/{id}"
        />
      </td>
      <td className="py-3 pr-2">
        <input
          value={path.description}
          onChange={(event) => onChange({ ...path, description: event.target.value })}
          className={buildInputClass(descriptionChanged, isAdded, 'w-full px-2 py-2 text-xs')}
          placeholder="path 설명 (선택)"
        />
      </td>
      <td className="py-3 pr-2">
        <NumberInput
          value={path.requestTimeout}
          min={1}
          isChanged={path.requestTimeout !== (baselinePath?.requestTimeout ?? path.requestTimeout)}
          isAdded={isAdded}
          onChange={(value) => onChange({ ...path, requestTimeout: value })}
        />
      </td>
      <td className="py-3 pr-2">
        <NumberInput
          value={path.responseTimeout}
          min={1}
          isChanged={path.responseTimeout !== (baselinePath?.responseTimeout ?? path.responseTimeout)}
          isAdded={isAdded}
          onChange={(value) => onChange({ ...path, responseTimeout: value })}
        />
      </td>
      <td className="py-3 pr-2">
        <NumberInput
          value={path.cacheTimeout}
          isChanged={path.cacheTimeout !== (baselinePath?.cacheTimeout ?? path.cacheTimeout)}
          isAdded={isAdded}
          onChange={(value) => onChange({ ...path, cacheTimeout: value })}
        />
      </td>
      <td className="py-3 pr-2">
        <div className="flex items-center gap-2">
          <label className="inline-flex items-center justify-center">
            <input
              type="checkbox"
              checked={path.useRateLimit}
              onChange={(event) =>
                onChange({
                  ...path,
                  useRateLimit: event.target.checked,
                  rateLimitCount: event.target.checked ? Math.max(1, path.rateLimitCount) : 0,
                })
              }
            />
          </label>
          <NumberInput
            value={path.rateLimitCount}
            min={path.useRateLimit ? 1 : 0}
            isChanged={path.rateLimitCount !== (baselinePath?.rateLimitCount ?? path.rateLimitCount)}
            isAdded={isAdded}
            disabled={!path.useRateLimit}
            onChange={(value) =>
              onChange({
                ...path,
                rateLimitCount: path.useRateLimit ? Math.max(1, value) : 0,
              })
            }
          />
        </div>
      </td>
      <td className="py-3 pr-2">
        <label className="inline-flex items-center justify-center">
          <input
            type="checkbox"
            checked={path.checkAuthorization}
            onChange={(event) => onChange({ ...path, checkAuthorization: event.target.checked })}
          />
        </label>
      </td>
      <td className="py-3">
        <button
          type="button"
          onClick={onRemove}
          className="rounded-md border border-slate-700 px-2 py-2 text-xs text-slate-200 transition hover:border-rose-500 hover:text-rose-200"
        >
          제거
        </button>
      </td>
    </tr>
  );
}

function NumberInput({
  value,
  onChange,
  min = 0,
  isChanged = false,
  isAdded = false,
  disabled = false,
}: {
  value: number;
  onChange: (next: number) => void;
  min?: number;
  isChanged?: boolean;
  isAdded?: boolean;
  disabled?: boolean;
}) {
  return (
    <input
      type="number"
      min={min}
      value={value}
      onChange={(event) => {
        const parsed = Number(event.target.value);
        const safeValue = Number.isNaN(parsed) ? min : Math.max(min, Math.trunc(parsed));
        onChange(safeValue);
      }}
      disabled={disabled}
      className={`${buildInputClass(isChanged, isAdded, 'w-full px-2 py-2 text-xs')} disabled:cursor-not-allowed disabled:opacity-60`}
    />
  );
}

function findResourceByDomain(resources: UpstreamResourceDraft[], domain: string): UpstreamResourceDraft | undefined {
  return resources.find((resource) => resource.domain === domain);
}

function isFieldChanged(current: string, baseline?: string): boolean {
  if (typeof baseline !== 'string') {
    return false;
  }
  return current !== baseline;
}

function buildInputClass(isChanged: boolean, isAdded = false, sizeClass = 'w-full px-3 py-2 text-sm'): string {
  if (isAdded) {
    return `${sizeClass} rounded-md border border-emerald-500/60 bg-emerald-500/10 text-white`;
  }
  if (isChanged) {
    return `${sizeClass} rounded-md border border-amber-500/70 bg-amber-500/10 text-white`;
  }
  return `${sizeClass} rounded-md border border-slate-700 bg-slate-950 text-white`;
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

function Card({
  title,
  children,
  action,
}: {
  title: string;
  children: React.ReactNode;
  action?: React.ReactNode;
}) {
  return (
    <div className="rounded-2xl border border-slate-800 bg-slate-900/60 p-5 shadow-xl shadow-slate-950/20">
      <div className="mb-4 flex items-center justify-between gap-3">
        <h3 className="text-lg font-semibold text-white">{title}</h3>
        {action}
      </div>
      {children}
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="mb-2 block text-sm font-medium text-slate-200">{label}</span>
      {children}
    </label>
  );
}
