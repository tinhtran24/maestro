import { Link2, Network } from "lucide-react";
import type { MemoryNode } from "../../domain/models";

export function MemoryDetail({
  node,
  related,
  onSelect,
}: {
  node: MemoryNode | null;
  related: MemoryNode[];
  onSelect: (id: string) => void;
}) {
  if (!node) {
    return (
      <div className="grid h-full place-items-center rounded-xl border border-dashed border-slate-700 bg-slate-950/40 p-6 text-center text-sm text-text-muted">
        Select a memory node to view its details and relationships.
      </div>
    );
  }
  return (
    <div className="grid gap-4 rounded-xl border border-slate-800 bg-slate-900/70 p-4">
      <div>
        <div className="flex items-center gap-2">
          <span className="rounded-md border border-blue-info/30 bg-blue-info/10 px-2 py-0.5 text-[11px] capitalize text-blue-info">{node.type}</span>
          <h3 className="text-lg font-semibold text-text-main">{node.title}</h3>
        </div>
        <p className="mt-3 text-sm text-text-main">{node.content}</p>
      </div>

      <div>
        <h4 className="mb-2 inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-text-muted">
          <Network size={13} /> Related memory
        </h4>
        {related.length ? (
          <ul className="grid gap-1.5">
            {related.map((item) => (
              <li key={item.id}>
                <button onClick={() => onSelect(item.id)} className="flex w-full items-center gap-2 rounded-lg border border-slate-800 bg-bg-card px-3 py-2 text-left text-sm hover:border-slate-700">
                  <Link2 size={13} className="shrink-0 text-text-muted" />
                  <span className="truncate text-text-main">{item.title}</span>
                  <span className="ml-auto shrink-0 text-[11px] capitalize text-text-muted">{item.type}</span>
                </button>
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-xs text-text-muted">No related memory found.</p>
        )}
      </div>
    </div>
  );
}
