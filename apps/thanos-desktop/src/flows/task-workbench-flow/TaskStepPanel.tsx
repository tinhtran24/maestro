import { Terminal } from "lucide-react";
import type { AgentSession, Task, WorkflowStepConfig } from "../../domain/models";

export function TaskStepPanel({ task, step, session, onOpenTerminal }: { task: Task; step: WorkflowStepConfig; session: AgentSession; onOpenTerminal: () => void }) {
  return (
    <article className="rounded-lg border border-slate-800 bg-bg-card p-3">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h4 className="font-semibold">Step Settings</h4>
          <p className="mt-1 text-xs text-text-muted">{task.status} · {step.label}</p>
        </div>
        <button onClick={onOpenTerminal} className="rounded-md border border-slate-700 p-2 text-text-muted"><Terminal size={15} /></button>
      </div>
      <dl className="mt-3 grid gap-2 text-xs">
        <div className="flex justify-between gap-3"><dt className="text-text-muted">Agent</dt><dd>{step.provider}</dd></div>
        <div className="flex justify-between gap-3"><dt className="text-text-muted">Command</dt><dd className="truncate">{step.command}</dd></div>
        <div className="flex justify-between gap-3"><dt className="text-text-muted">Terminal</dt><dd>{session.status}</dd></div>
        <div className="flex justify-between gap-3"><dt className="text-text-muted">Approval</dt><dd>{step.approvalRequired ? "Required" : "Not required"}</dd></div>
      </dl>
    </article>
  );
}
