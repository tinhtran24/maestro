import { Brain, ClipboardList, Code2, FlaskConical, GitPullRequest } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { Task, WorkflowStepId } from "../../domain/models";
import { Card } from "../../shared/ui/Card";
import { currentWorkflowStep } from "../../state/workbenchStore";

const steps: Array<{ id: WorkflowStepId; label: string; icon: LucideIcon }> = [
  { id: "planning", label: "Planning", icon: ClipboardList },
  { id: "coding", label: "Coding", icon: Code2 },
  { id: "review", label: "Review", icon: GitPullRequest },
  { id: "testing", label: "Testing", icon: FlaskConical },
  { id: "memory_update", label: "Memory Update", icon: Brain },
];

const order: WorkflowStepId[] = ["planning", "coding", "review", "testing", "memory_update"];

type Phase = "done" | "active" | "pending";

const badge: Record<Phase, { label: string; className: string }> = {
  done: { label: "Done", className: "border-green-success/30 bg-green-success/10 text-green-success" },
  active: { label: "Running", className: "border-blue-info/30 bg-blue-info/10 text-blue-info" },
  pending: { label: "Pending", className: "border-slate-700 bg-slate-800/60 text-text-muted" },
};

const statusText: Record<Phase, string> = { done: "Completed", active: "In progress", pending: "Not started" };

function phaseFor(stepId: WorkflowStepId, task: Task | null): Phase {
  if (!task) return "pending";
  const current = currentWorkflowStep(task);
  const currentIndex = order.indexOf(current === "debugging" || current === "documentation" ? "coding" : current);
  const stepIndex = order.indexOf(stepId);
  if (task.status === "done") return "done";
  if (stepIndex < currentIndex) return "done";
  if (stepIndex === currentIndex) return "active";
  return "pending";
}

export function ActiveWorkflowCard({ task }: { task: Task | null }) {
  return (
    <Card title="Active Workflow Steps">
      <ul className="grid gap-1">
        {steps.map(({ id, label, icon: Icon }) => {
          const phase = phaseFor(id, task);
          const tone = badge[phase];
          return (
            <li key={id} className="flex items-center gap-3 rounded-lg px-2 py-2 hover:bg-slate-800/50">
              <span className={`grid h-7 w-7 shrink-0 place-items-center rounded-md border border-slate-800 bg-slate-950 ${phase === "active" ? "text-blue-info" : phase === "done" ? "text-green-success" : "text-text-muted"}`}>
                <Icon size={14} />
              </span>
              <span className="min-w-0 flex-1">
                <span className="block truncate text-sm font-medium text-text-main">{label}</span>
                <span className="block truncate text-xs text-text-muted">{statusText[phase]}</span>
              </span>
              <span className={`shrink-0 rounded-md border px-2 py-0.5 text-[11px] font-medium ${tone.className}`}>{tone.label}</span>
            </li>
          );
        })}
      </ul>
    </Card>
  );
}
