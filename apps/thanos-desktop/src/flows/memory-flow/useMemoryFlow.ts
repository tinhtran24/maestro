// Phase 9 — Memory (mock).
// Orchestrates the memory browser: full-text search (mock FTS), type filtering,
// node selection, and related-node derivation. Backed by the in-memory store —
// no real SQLite yet.

import { useMemo, useState } from "react";
import type { MemoryNode } from "../../domain/models";
import { relatedMemory, searchMemory } from "../../features/memory/memorySearch";
import { useWorkbenchStore } from "../../state/workbenchStore";

export type MemoryTypeFilter = MemoryNode["type"] | "all";

export function useMemoryFlow() {
  const nodes = useWorkbenchStore((state) => state.memoryNodes);
  const [query, setQuery] = useState("");
  const [type, setType] = useState<MemoryTypeFilter>("all");
  const [selectedId, setSelectedId] = useState<string>("");

  const types = useMemo(() => Array.from(new Set(nodes.map((node) => node.type))), [nodes]);

  const results = useMemo(() => {
    const matched = searchMemory(nodes, query);
    return type === "all" ? matched : matched.filter((node) => node.type === type);
  }, [nodes, query, type]);

  const selected = results.find((node) => node.id === selectedId) ?? results[0] ?? null;
  const related = selected ? relatedMemory(nodes, selected) : [];

  return {
    query,
    setQuery,
    type,
    setType,
    types,
    results,
    selected,
    related,
    select: setSelectedId,
    total: nodes.length,
  };
}
