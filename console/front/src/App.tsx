import { useCallback, useEffect, useState } from 'react';
import { deleteService, getService, listServices, createService, updateService, upsertServiceResource } from './lib/api';
import { createServiceDraft } from './lib/defaults';
import { normalizeDraftFromApi } from './lib/mappers';
import { validateDraft } from './lib/validation';
import { ServiceEditor } from './components/ServiceEditor';
import { ServiceList } from './components/ServiceList';
import type { UpstreamResourceDraft, UpstreamServiceDraft, UpstreamServiceSummary } from './types';

export default function App() {
  const [services, setServices] = useState<UpstreamServiceSummary[]>([]);
  const [selectedServiceName, setSelectedServiceName] = useState<string | null>(null);
  const [draft, setDraft] = useState<UpstreamServiceDraft>(createServiceDraft());
  const [baselineDraft, setBaselineDraft] = useState<UpstreamServiceDraft | null>(null);
  const [mode, setMode] = useState<'create' | 'edit'>('create');
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const loadServices = useCallback(async (preferredServiceName?: string | null) => {
    const nextServices = await listServices();
    setServices(nextServices);

    if (nextServices.length === 0) {
      setSelectedServiceName(null);
      setDraft(createServiceDraft());
      setBaselineDraft(null);
      setMode('create');
      return;
    }

    const nextSelection = preferredServiceName ?? selectedServiceName;
    if (nextSelection && nextServices.some((service) => service.serviceName === nextSelection)) {
      setSelectedServiceName(nextSelection);
      return;
    }

    setSelectedServiceName(nextServices[0].serviceName);
  }, [selectedServiceName]);

  useEffect(() => {
    void loadServices().catch((loadError: Error) => {
      setError(loadError.message);
    });
  }, [loadServices]);

  useEffect(() => {
    if (!selectedServiceName) {
      return;
    }

    let cancelled = false;
    setDraft(createServiceDraft());
    setMode('create');
    setBusy(true);
    setError(null);
    void getService(selectedServiceName)
      .then((service) => {
        if (cancelled) {
          return;
        }
        const normalized = normalizeDraftFromApi(service);
        setDraft(normalized);
        setBaselineDraft(normalized);
        setMode('edit');
      })
      .catch((loadError: Error) => {
        if (cancelled) {
          return;
        }
        setDraft(createServiceDraft());
        setBaselineDraft(null);
        setMode('create');
        setError(loadError.message);
      })
      .finally(() => {
        if (cancelled) {
          return;
        }
        setBusy(false);
      });

    return () => {
      cancelled = true;
    };
  }, [selectedServiceName]);

  const handleCreate = () => {
    setSelectedServiceName(null);
    setDraft(createServiceDraft());
    setBaselineDraft(null);
    setMode('create');
    setMessage(null);
    setError(null);
  };

  const handleSave = async () => {
    const validationErrors = validateDraft(draft);
    if (validationErrors.length > 0) {
      setError(validationErrors[0]);
      return;
    }

    setBusy(true);
    setMessage(null);
    setError(null);

    try {
      if (mode === 'create') {
        const created = await createService(draft);
        const normalized = normalizeDraftFromApi(created);
        setSelectedServiceName(created.serviceName);
        setDraft(normalized);
        setBaselineDraft(normalized);
        setMode('edit');
        setMessage('규칙을 생성했습니다.');
        await loadServices(created.serviceName);
      } else {
        const updated = await updateService(draft.serviceName, draft);
        const normalized = normalizeDraftFromApi(updated);
        setDraft(normalized);
        setBaselineDraft(normalized);
        setMessage('규칙을 저장했습니다.');
        await loadServices(updated.serviceName);
      }
    } catch (saveError) {
      setError(saveError instanceof Error ? saveError.message : '저장에 실패했습니다.');
    } finally {
      setBusy(false);
    }
  };

  const handleDelete = async () => {
    if (mode !== 'edit') {
      return;
    }

    const confirmed = window.confirm(`service "${draft.serviceName}" 규칙을 삭제할까요?`);
    if (!confirmed) {
      return;
    }

    setBusy(true);
    setMessage(null);
    setError(null);
    try {
      await deleteService(draft.serviceName);
      setMessage('규칙을 삭제했습니다.');
      setSelectedServiceName(null);
      setDraft(createServiceDraft());
      setBaselineDraft(null);
      setMode('create');
      await loadServices(null);
    } catch (deleteError) {
      setError(deleteError instanceof Error ? deleteError.message : '삭제에 실패했습니다.');
    } finally {
      setBusy(false);
    }
  };

  const handleSaveResource = async (resource: UpstreamResourceDraft, previousDomain: string) => {
    if (mode !== 'edit') {
      await handleSave();
      return;
    }

    setBusy(true);
    setMessage(null);
    setError(null);
    try {
      const updated = await upsertServiceResource(draft.serviceName, resource, previousDomain);
      const normalized = normalizeDraftFromApi(updated);
      setDraft(normalized);
      setBaselineDraft(normalized);
      setMessage('domain 규칙을 저장했습니다.');
      await loadServices(updated.serviceName);
    } catch (saveError) {
      setError(saveError instanceof Error ? saveError.message : 'domain 저장에 실패했습니다.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <main className="flex min-h-screen bg-slate-950 text-slate-100">
      <ServiceList
        services={services}
        selectedServiceName={selectedServiceName}
        onSelect={setSelectedServiceName}
        onCreate={handleCreate}
      />
      <ServiceEditor
        draft={draft}
        mode={mode}
        busy={busy}
        message={message}
        error={error}
        baselineDraft={baselineDraft}
        onChange={setDraft}
        onSaveResource={(resource, previousDomain) => void handleSaveResource(resource, previousDomain)}
        onDelete={() => void handleDelete()}
      />
    </main>
  );
}
