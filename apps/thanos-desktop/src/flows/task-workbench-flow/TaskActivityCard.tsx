import { ArrowRight, Ban, Check, CheckCircle2, FlaskConical, ListTodo, Play, RotateCcw, ShieldCheck, XCircle } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { Task, TaskEvent, TaskEventType } from "../../domain/models";
import { Card } from "../../shared/ui/Card";
import { statusLabel } from "../../state/taskMachine";
import { taskEventsFor, useWorkbenchStore } from "../../state/workbenchStore";

const meta: Record<TaskEventType, { icon: LucideIcon; tone: string; label: (event: TaskEvent) => string }> = {
  created: { icon: ListTodo, tone: "text-text-muted", label: () => "Task created" },
  moved: { icon: ArrowRight, tone: "text-blue-info", label: (e) => `Moved to ${e.to ? statusLabel(e.to) : "next"}` },
  plan_approved: { icon: Check, tone: "text-green-success", label: () => "Plan approved → Ready" },
  changes_requested: { icon: RotateCcw, tone: "text-yellow-warning", label: () => "Changes requested" },
  agent_started: { icon: Play, tone: "text-blue-info", label: () => "Agent started (In Progress)" },
  tests_passed: { icon: FlaskConical, tone: "text-green-success", label: () => "Tests passed" },
  review_approved: { icon: ShieldCheck, tone: "text-green-success", label: () => "Review approved" },
  blocked: { icon: Ban, tone: "text-yellow-warning", label: () => "Blocked" },
  failed: { icon: XCircle, tone: "text-red-danger", label: () => "Failed" },
  done: { icon: CheckCircle2, tone: "text-green-success", label: () => "Done" },
};

function formatTime(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleString([], { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" });
}

export function TaskActivityCard({ task }: { task: Task }) {
  const events = useWorkbenchStore((state) => taskEventsFor(state, task.id));
  const ordered = events.slice().reverse();

  return (
    <Card title="Recent Activity" action={<span className="text-xs text-text-muted">{events.length}</span>}>
      {ordered.length === 0 ? (
        <p className="rounded-lg border border-slate-800 bg-bg-card p-3 text-xs text-text-muted">No workflow activity yet.</p>
      ) : (
        <ol className="grid gap-0">
          {ordered.map((event, index) => {
            const entry = meta[event.type];
            const Icon = entry.icon;
            return (
              <li key={event.id} className="flex gap-3">
                <div className="flex flex-col items-center">
                  <span className={`grid h-6 w-6 shrink-0 place-items-center rounded-full border border-slate-800 bg-bg-card ${entry.tone}`}>
                    <Icon size={12} />
                  </span>
                  {index < ordered.length - 1 && <span className="my-0.5 w-px flex-1 bg-slate-800" style={{ minHeight: 12 }} />}
                </div>
                <div className="min-w-0 pb-3">
                  <p className="text-sm text-text-main">{entry.label(event)}</p>
                  <p className="mt-0.5 text-[11px] text-text-muted">{formatTime(event.at)}{event.note ? ` · ${event.note}` : ""}</p>
                </div>
              </li>
            );
          })}
        </ol>
      )}
    </Card>
  );
}
