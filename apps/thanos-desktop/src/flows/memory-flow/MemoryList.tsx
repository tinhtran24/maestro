import type { MemoryNode } from "../../domain/models";

export function MemoryList({
  nodes,
  selectedId,
  onSelect,
}: {
  nodes: MemoryNode[];
  selectedId?: string;
  onSelect: (id: string) => void;
}) {
  if (!nodes.length) {
    return <p className="rounded-lg border border-slate-800 bg-bg-card p-3 text-xs text-text-muted">No memory matches your search.</p>;
  }
  return (
    <ul className="grid gap-2">
      {nodes.map((node) => (
        <li key={node.id}>
          <button
            onClick={() => onSelect(node.id)}
            className={`w-full rounded-lg border p-3 text-left transition ${node.id === selectedId ? "border-purple-primary bg-purple-primary/10" : "border-slate-800 bg-slate-900/70 hover:border-slate-700"}`}
          >
            <div className="flex items-center justify-between gap-2">
              <span className="truncate text-sm font-medium text-text-main">{node.title}</span>
              <span className="shrink-0 rounded-md border border-blue-info/30 bg-blue-info/10 px-2 py-0.5 text-[11px] capitalize text-blue-info">{node.type}</span>
            </div>
            <p className="mt-1 line-clamp-2 text-xs text-text-muted">{node.content}</p>
          </button>
        </li>
      ))}
    </ul>
  );
}
