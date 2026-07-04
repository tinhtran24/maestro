// Phase 9 — Memory.
// Pure memory helpers. Phase 9 is UI-first: there is no real SQLite/FTS backend
// yet, so search is a mock full-text filter over the in-memory nodes and related
// nodes are derived from explicit links + shared type. Memory synthesis turns an
// approved task/plan into a decision node ("memory updates after approval").

import type { ExecutionPlan, MemoryNode, Review, Task } from "../../domain/models";

const STOPWORDS = new Set(["the", "a", "an", "and", "or", "of", "to", "for", "in", "on", "is", "are"]);

function tokenize(value: string): string[] {
  return value
    .toLowerCase()
    .split(/[^a-z0-9]+/)
    .filter((token) => token.length > 1 && !STOPWORDS.has(token));
}

// Mock full-text search: every query token must appear in a node's searchable
// text (title + content + type). An empty query returns all nodes unchanged.
export function searchMemory(nodes: MemoryNode[], query: string): MemoryNode[] {
  const tokens = tokenize(query);
  if (tokens.length === 0) return nodes;
  return nodes.filter((node) => {
    const haystack = `${node.title} ${node.content} ${node.type}`.toLowerCase();
    return tokens.every((token) => haystack.includes(token));
  });
}

// Related nodes: explicit links first (in either direction), then others of the
// same type. The node itself is always excluded.
export function relatedMemory(nodes: MemoryNode[], node: MemoryNode): MemoryNode[] {
  const linked = new Set(node.links);
  const others = nodes.filter((item) => item.id !== node.id);
  const byLink = others.filter((item) => linked.has(item.id) || item.links.includes(node.id));
  const linkedIds = new Set(byLink.map((item) => item.id));
  const byType = others.filter((item) => !linkedIds.has(item.id) && item.type === node.type);
  return [...byLink, ...byType];
}

// Synthesizes a decision memory node from an approved task + its plan. Links back
// to any prior memory whose title token-overlaps the task, forming relationships.
export function memoryFromReview(task: Task, plan: ExecutionPlan, existing: MemoryNode[], createdAt: string, review?: Review): MemoryNode {
  const taskTokens = new Set(tokenize(task.title));
  const links = existing
    .filter((node) => tokenize(node.title).some((token) => taskTokens.has(token)))
    .map((node) => node.id);
  const notes = review?.reviewerNotes?.trim();
  const content = [
    plan.summary,
    notes ? `Review: ${notes}` : "",
    plan.filesToTouch.length ? `Files: ${plan.filesToTouch.join(", ")}` : "",
  ]
    .filter(Boolean)
    .join(" ");

  return {
    id: `mem-${task.id.toLowerCase()}`,
    projectId: task.featureId,
    type: "decision",
    title: `${task.title} — approved`,
    content: content || `Approved implementation of ${task.title}.`,
    links,
    createdAt,
  };
}
