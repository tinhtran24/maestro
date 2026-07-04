import { AlertTriangle, CheckCircle2, Code2, Play } from "lucide-react";
import type { Task } from "../../domain/models";
import { CodingChangeset } from "./CodingChangeset";
import { CodingReviewGate } from "./CodingReviewGate";
import { WorktreeCard } from "./WorktreeCard";
import { useCoderFlow } from "./useCoderFlow";

const CODING_STATUSES = ["ready", "running"];

// Shown at the top of the task workbench while a task is in the coding phase.
export function isCodingPhase(task: Task): boolean {
  return CODING_STATUSES.includes(task.status);
}

export function CoderFlow({ task }: { task: Task }) {
  const flow = useCoderFlow(task);

  return (
    <section className="rounded-xl border border-slate-800 bg-slate-900/70 p-4 shadow-lg shadow-black/20">
      <header className="mb-3 flex items-center justify-between gap-2">
        <h3 className="inline-flex items-center gap-2 text-base font-semibold">
          <Code2 size={16} className="text-purple-hover" /> Coder
        </h3>
        <span className="text-xs text-text-muted">Coding · {task.status.replace("_", " ")}</span>
      </header>

      {flow.stage === "blocked" && (
        <p className="inline-flex items-center gap-2 text-sm text-yellow-warning">
          <AlertTriangle size={15} /> Coding is blocked until the execution plan is approved.
        </p>
      )}

      {flow.stage === "ready" && (
        <div className="grid gap-3">
          <p className="text-sm text-text-muted">The plan is approved. Create an isolated worktree and open the coding terminal to apply the planned changes.</p>
          <div>
            <button onClick={flow.startCoding} className="inline-flex items-center gap-2 rounded-lg bg-purple-primary px-3 py-2 text-sm font-medium text-white hover:bg-purple-hover">
              <Play size={15} /> Create Worktree & Start Coding
            </button>
          </div>
        </div>
      )}

      {(flow.stage === "coding" || flow.stage === "review") && (
        <div className="grid gap-4">
          <WorktreeCard task={task} />
          <CodingChangeset diff={flow.diff} />
          <CodingReviewGate ready={flow.codingComplete} onSendToReview={flow.sendToReview} />
        </div>
      )}

      {flow.stage === "done" && (
        <div className="grid gap-4">
          <p className="inline-flex items-center gap-2 text-sm text-green-success">
            <CheckCircle2 size={15} /> Coding complete — task moved to review.
          </p>
          <WorktreeCard task={task} />
          <CodingChangeset diff={flow.diff} />
        </div>
      )}
    </section>
  );
}
