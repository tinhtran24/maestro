import { Search } from "lucide-react";
import type { MemoryNode } from "../../domain/models";
import type { MemoryTypeFilter } from "./useMemoryFlow";

export function MemorySearch({
  query,
  onQuery,
  type,
  onType,
  types,
}: {
  query: string;
  onQuery: (value: string) => void;
  type: MemoryTypeFilter;
  onType: (value: MemoryTypeFilter) => void;
  types: MemoryNode["type"][];
}) {
  const filters: MemoryTypeFilter[] = ["all", ...types];
  return (
    <div className="grid gap-3">
      <div className="relative">
        <Search size={15} className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-text-muted" />
        <input
          value={query}
          onChange={(event) => onQuery(event.target.value)}
          placeholder="Search project memory…"
          className="h-10 w-full rounded-lg border border-slate-800 bg-slate-950/60 pl-9 pr-3 text-sm text-text-main placeholder:text-text-muted focus:border-slate-600 focus:outline-none"
        />
      </div>
      <div className="flex flex-wrap gap-1.5">
        {filters.map((filter) => (
          <button
            key={filter}
            onClick={() => onType(filter)}
            className={`rounded-lg border px-2.5 py-1 text-xs capitalize ${type === filter ? "border-purple-primary bg-purple-primary/15 text-purple-hover" : "border-slate-800 bg-slate-900/70 text-text-muted hover:border-slate-700"}`}
          >
            {filter}
          </button>
        ))}
      </div>
    </div>
  );
}
