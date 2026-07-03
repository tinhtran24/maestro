import { Check } from "lucide-react";
import type { Task, TaskStatus } from "../../domain/models";
import { Card } from "../../shared/ui/Card";

const stepper: Array<{ id: string; label: string }> = [
  { id: "created", label: "Ticket Created" },
  { id: "planning", label: "Planning" },
  { id: "waiting_approval", label: "Waiting Approval" },
  { id: "coding", label: "Coding" },
  { id: "review", label: "Review" },
  { id: "testing", label: "Testing" },
  { id: "done", label: "Done" },
];

const currentIndexFor: Record<TaskStatus, number> = {
  backlog: 1,
  planning: 1,
  waiting_approval: 2,
  ready: 3,
  running: 3,
  in_review: 4,
  waiting_user: 4,
  blocked: 3,
  failed: 3,
  done: 6,
};

export function WorkflowTimelineCard({ task }: { task: Task }) {
  const current = task.status === "done" ? stepper.length : currentIndexFor[task.status];
  return (
    <Card title="Workflow Timeline">
      <ol className="grid gap-0">
        {stepper.map((step, index) => {
          const done = index < current;
          const active = index === current;
          return (
            <li key={step.id} className="flex gap-3">
              <div className="flex flex-col items-center">
                <span className={`grid h-6 w-6 shrink-0 place-items-center rounded-full border text-[11px] ${done ? "border-green-success bg-green-success/15 text-green-success" : active ? "border-blue-info bg-blue-info/15 text-blue-info" : "border-slate-700 bg-slate-950 text-text-muted"}`}>
                  {done ? <Check size={12} /> : active ? <span className="h-2 w-2 rounded-full bg-blue-info" /> : <span className="h-1.5 w-1.5 rounded-full bg-slate-600" />}
                </span>
                {index < stepper.length - 1 && <span className={`my-0.5 w-px flex-1 ${done ? "bg-green-success/40" : "bg-slate-800"}`} style={{ minHeight: 14 }} />}
              </div>
              <span className={`pb-3 pt-0.5 text-sm ${active ? "font-medium text-text-main" : done ? "text-text-main" : "text-text-muted"}`}>{step.label}</span>
            </li>
          );
        })}
      </ol>
    </Card>
  );
}
