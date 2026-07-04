// Phase 7 — Coding Workflow.
// Pure, mock changeset generation. Phase 7 is UI-first: the "coder" does not run
// a real agent or touch Git. It synthesizes a deterministic changeset (changed
// files with add/modify status, insertion/deletion counts, and a unified-diff-ish
// patch) from the task and its approved execution plan. No execution, no Git.

import type { ExecutionPlan, GitDiff, Task } from "../../domain/models";

// Deterministic 32-bit hash so line counts are stable across runs (tests rely on
// this — the mock coder must never depend on Math.random or the clock).
function hash(value: string): number {
  let h = 2166136261;
  for (let i = 0; i < value.length; i += 1) {
    h ^= value.charCodeAt(i);
    h = Math.imul(h, 16777619);
  }
  return h >>> 0;
}

// A file is "added" when it has no extension yet or ends with a path separator
// (a directory placeholder like "src/"); otherwise it is "modified".
function statusFor(path: string): "added" | "modified" {
  const isDirectory = path.endsWith("/");
  const hasExtension = /\.[a-z0-9]+$/i.test(path);
  return isDirectory || !hasExtension ? "added" : "modified";
}

type FileChange = { path: string; status: "added" | "modified"; additions: number; deletions: number };

function changesFor(files: string[]): FileChange[] {
  return files.map((path) => {
    const status = statusFor(path);
    const seed = hash(path);
    const additions = 6 + (seed % 40);
    // Added files only insert; modified files also remove a smaller amount.
    const deletions = status === "added" ? 0 : 1 + ((seed >> 8) % 12);
    return { path, status, additions, deletions };
  });
}

function buildPatch(task: Task, changes: FileChange[]): string {
  const header = `# Coding changeset for ${task.id} — ${task.title}`;
  const hunks = changes.map((change) => {
    const marker = change.status === "added" ? "new file" : "modified";
    return [
      `diff --git a/${change.path} b/${change.path}`,
      `${marker}  +${change.additions} -${change.deletions}`,
      change.status === "added"
        ? `+ // ${change.path}: implemented per approved execution plan`
        : `@@ ${change.path} @@ applied planned changes`,
    ].join("\n");
  });
  return [header, "", ...hunks].join("\n");
}

// Synthesizes the coder's changeset from the task and its approved plan.
export function generateChangeset(task: Task, plan: ExecutionPlan): GitDiff {
  const files = plan.filesToTouch.length ? plan.filesToTouch : ["src/"];
  const changes = changesFor(files);
  const additions = changes.reduce((total, change) => total + change.additions, 0);
  const deletions = changes.reduce((total, change) => total + change.deletions, 0);
  const summary = `${changes.length} file${changes.length === 1 ? "" : "s"} changed, ${additions} insertion${additions === 1 ? "" : "s"}(+), ${deletions} deletion${deletions === 1 ? "" : "s"}(-)`;

  return {
    taskId: task.id,
    summary,
    changedFiles: changes.map((change) => ({ path: change.path, status: change.status })),
    patch: buildPatch(task, changes),
  };
}
