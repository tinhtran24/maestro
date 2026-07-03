import { CheckCircle2, Play, ScrollText } from "lucide-react";
import type { Task } from "../../domain/models";
import { ExecutionPlanPreview } from "./ExecutionPlanPreview";
import { PlanApprovalGate } from "./PlanApprovalGate";
import { PlanningQuestions } from "./PlanningQuestions";
import { usePlannerFlow } from "./usePlannerFlow";

const PLANNING_STATUSES = ["backlog", "planning", "waiting_approval"];

// Shown at the top of the task workbench while a task is in the planning phase.
export function isPlanningPhase(task: Task): boolean {
  return PLANNING_STATUSES.includes(task.status);
}

export function PlannerFlow({ task }: { task: Task }) {
  const flow = usePlannerFlow(task);

  return (
    <section className="rounded-xl border border-slate-800 bg-slate-900/70 p-4 shadow-lg shadow-black/20">
      <header className="mb-3 flex items-center justify-between gap-2">
        <h3 className="inline-flex items-center gap-2 text-base font-semibold">
          <ScrollText size={16} className="text-purple-hover" /> Planner
        </h3>
        <span className="text-xs text-text-muted">Planning · {task.status.replace("_", " ")}</span>
      </header>

      {flow.stage === "start" && (
        <div className="grid gap-3">
          <p className="text-sm text-text-muted">Start planning to open a Claude terminal and answer a few clarifying questions before an execution plan is drafted.</p>
          <div>
            <button onClick={flow.startPlanning} className="inline-flex items-center gap-2 rounded-lg bg-purple-primary px-3 py-2 text-sm font-medium text-white hover:bg-purple-hover">
              <Play size={15} /> Start Planning Agent
            </button>
          </div>
        </div>
      )}

      {flow.stage === "questions" && (
        <PlanningQuestions questions={flow.questions} onAnswer={flow.answer} onSubmit={flow.submitPlan} canSubmit={flow.answered} />
      )}

      {flow.stage === "review" && (
        <div className="grid gap-4">
          <PlanApprovalGate onApprove={flow.approve} onReject={flow.reject} />
          <ExecutionPlanPreview plan={flow.plan} />
        </div>
      )}

      {flow.stage === "approved" && (
        <div className="grid gap-4">
          <p className="inline-flex items-center gap-2 text-sm text-green-success">
            <CheckCircle2 size={15} /> Plan approved — task is ready for coding.
          </p>
          <ExecutionPlanPreview plan={flow.plan} />
        </div>
      )}
    </section>
  );
}
