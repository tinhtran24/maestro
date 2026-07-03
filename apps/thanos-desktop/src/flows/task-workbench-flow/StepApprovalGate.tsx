import type { Task, WorkflowStepConfig } from "../../domain/models";

export function StepApprovalGate({ task, step }: { task: Task; step: WorkflowStepConfig }) {
  const blockedCoding = step.id === "coding" && !task.planApproved;
  return (
    <div className={`rounded-lg border p-3 text-xs ${blockedCoding ? "border-yellow-warning/40 bg-yellow-warning/10 text-yellow-warning" : "border-slate-800 bg-bg-card text-text-muted"}`}>
      {blockedCoding ? "Coding is blocked until the execution plan is approved." : step.approvalRequired ? "This step pauses before the next workflow transition." : "This step may continue after command completion."}
    </div>
  );
}
