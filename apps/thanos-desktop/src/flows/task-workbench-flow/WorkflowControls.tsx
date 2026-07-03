import { ArrowRight, Ban, RotateCcw, XCircle } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { Task, TaskStatus } from "../../domain/models";
import { Card } from "../../shared/ui/Card";
import { StatusBadge } from "../../shared/ui/StatusBadge";
import { availableTransitions, MAIN_FLOW, statusLabel } from "../../state/taskMachine";
import { useWorkbenchStore } from "../../state/workbenchStore";

function iconFor(status: Task["status"], to: TaskStatus): LucideIcon {
  if (to === "blocked") return Ban;
  if (to === "failed") return XCircle;
  if (to === "planning" && status !== "backlog") return RotateCcw;
  return ArrowRight;
}

export function WorkflowControls({ task }: { task: Task }) {
  const advanceTask = useWorkbenchStore((state) => state.advanceTask);
  const options = availableTransitions(task);
  const mainNext = MAIN_FLOW[MAIN_FLOW.indexOf(task.status) + 1];
  const blocked = options.filter((option) => !option.allowed);

  return (
    <Card title="Workflow" action={<StatusBadge status={task.status} />}>
      {options.length === 0 ? (
        <p className="rounded-lg border border-slate-800 bg-bg-card p-3 text-xs text-text-muted">This task is complete. No further transitions.</p>
      ) : (
        <div className="grid grid-cols-2 gap-2">
          {options.map((option) => {
            const Icon = iconFor(task.status, option.to);
            const primary = option.to === mainNext;
            const danger = option.to === "blocked" || option.to === "failed";
            const base = "inline-flex items-center justify-center gap-2 rounded-lg border px-2 py-2 text-xs font-medium transition disabled:cursor-not-allowed disabled:opacity-40";
            const tone = primary
              ? "border-purple-primary bg-purple-primary text-white enabled:hover:bg-purple-hover"
              : danger
                ? "border-red-danger/40 text-red-danger enabled:hover:bg-red-danger/10"
                : "border-slate-800 bg-slate-900/70 text-text-main enabled:hover:border-slate-700";
            return (
              <button
                key={option.to}
                type="button"
                disabled={!option.allowed}
                title={option.reason ?? `Move to ${statusLabel(option.to)}`}
                onClick={() => advanceTask(task.id, option.to)}
                className={`${base} ${tone}`}
              >
                <Icon size={13} /> {statusLabel(option.to)}
              </button>
            );
          })}
        </div>
      )}
      {blocked.length > 0 && (
        <ul className="mt-3 grid gap-1">
          {blocked.map((option) => (
            <li key={option.to} className="text-[11px] text-text-muted">
              <span className="text-yellow-warning">•</span> {statusLabel(option.to)}: {option.reason}
            </li>
          ))}
        </ul>
      )}
    </Card>
  );
}
