import type { UpstreamServiceSummary } from '../types';

type ServiceListProps = {
  services: UpstreamServiceSummary[];
  selectedServiceName: string | null;
  onSelect: (serviceName: string) => void;
  onCreate: () => void;
};

export function ServiceList({ services, selectedServiceName, onSelect, onCreate }: ServiceListProps) {
  return (
    <aside className="flex h-full w-full max-w-sm flex-col border-r border-slate-800 bg-slate-900/70">
      <div className="flex items-center justify-between border-b border-slate-800 px-5 py-4">
        <div>
          <h1 className="text-lg font-semibold text-white">Gateway Backoffice</h1>
          <p className="text-sm text-slate-400">service 규칙 목록</p>
        </div>
        <button
          type="button"
          onClick={onCreate}
          className="rounded-md bg-sky-500 px-3 py-2 text-sm font-medium text-slate-950 transition hover:bg-sky-400"
        >
          새 규칙
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-3">
        {services.length === 0 ? (
          <div className="rounded-lg border border-dashed border-slate-700 p-4 text-sm text-slate-400">
            등록된 service 가 없습니다.
          </div>
        ) : (
          <ul className="space-y-2">
            {services.map((service) => {
              const isSelected = service.serviceName === selectedServiceName;
              return (
                <li key={service.serviceName}>
                  <button
                    type="button"
                    onClick={() => onSelect(service.serviceName)}
                    className={`w-full rounded-lg border px-4 py-3 text-left transition ${
                      isSelected
                        ? 'border-sky-400 bg-sky-500/10 text-white'
                        : 'border-slate-800 bg-slate-950/60 text-slate-200 hover:border-slate-700'
                    }`}
                  >
                    <div className="font-medium">{service.serviceName}</div>
                    <div className="mt-1 text-xs text-slate-400">
                      resources {service.resourceCount} / paths {service.pathCount}
                    </div>
                  </button>
                </li>
              );
            })}
          </ul>
        )}
      </div>
    </aside>
  );
}
