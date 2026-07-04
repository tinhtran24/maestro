import { Brain } from "lucide-react";
import { MemoryDetail } from "./MemoryDetail";
import { MemoryList } from "./MemoryList";
import { MemorySearch } from "./MemorySearch";
import { useMemoryFlow } from "./useMemoryFlow";

// Phase 9 — Memory. Full-screen project memory browser: search, filter, list, and
// related-node detail. Rendered for the "memory" view.
export function MemoryFlow() {
  const flow = useMemoryFlow();

  return (
    <div className="flex h-full min-h-0 flex-col bg-bg-app">
      <header className="flex items-center justify-between gap-3 border-b border-slate-800 bg-slate-950/60 px-6 py-4">
        <div>
          <h1 className="inline-flex items-center gap-2 text-lg font-semibold"><Brain size={18} className="text-purple-hover" /> Project Memory</h1>
          <p className="mt-1 text-sm text-text-muted">Decisions, architecture, features, and reviews — searchable and linked.</p>
        </div>
        <span className="rounded-lg border border-slate-800 bg-slate-900/70 px-3 py-1 text-xs text-text-muted">{flow.results.length} / {flow.total} nodes</span>
      </header>

      <div className="grid min-h-0 flex-1 gap-4 overflow-hidden p-6 lg:grid-cols-[minmax(0,22rem)_minmax(0,1fr)]">
        <div className="flex min-h-0 flex-col gap-3">
          <MemorySearch query={flow.query} onQuery={flow.setQuery} type={flow.type} onType={flow.setType} types={flow.types} />
          <div className="min-h-0 flex-1 overflow-y-auto pr-1">
            <MemoryList nodes={flow.results} selectedId={flow.selected?.id} onSelect={flow.select} />
          </div>
        </div>
        <div className="min-h-0 overflow-y-auto">
          <MemoryDetail node={flow.selected} related={flow.related} onSelect={flow.select} />
        </div>
      </div>
    </div>
  );
}
