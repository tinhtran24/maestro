import { AlertTriangle, ArrowRight, Ban, CheckCircle2, Circle, FileText, FlaskConical, GitCompare, ListTodo, Play, RotateCcw, ScrollText, ShieldCheck, Terminal, XCircle } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { Task, TaskEvent, TaskEventType } from "../../domain/models";
import { EmptyState } from "../../shared/ui/EmptyState";
import { statusLabel } from "../../state/taskMachine";
import { planFor, reviewFor, taskEventsFor, useWorkbenchStore } from "../../state/workbenchStore";

type TimelineEntry = {
  id: string;
  icon: LucideIcon;
  label: string;
  detail?: string;
  tone: "done" | "active" | "pending" | "warn";
};

const toneClass: Record<TimelineEntry["tone"], string> = {
  done: "text-green-success",
  active: "text-blue-info",
  pending: "text-slate-500",
  warn: "text-yellow-warning",
};

function buildTimeline(task: Task, state: ReturnType<typeof useWorkbenchStore.getState>): TimelineEntry[] {
  const plan = planFor(task, state);
  const review = reviewFor(task, state);
  const sessions = state.sessions.filter((session) => session.taskId === task.id);
  const diff = state.diffs[task.id];
  const test = state.testRuns[task.id];
  const entries: TimelineEntry[] = [];

  entries.push({
    id: "created",
    icon: ListTodo,
    label: `Task ${task.id} created`,
    detail: task.title,
    tone: "done",
  });

  entries.push({
    id: "plan",
    icon: FileText,
    label: `Execution plan · ${plan.approvalStatus.replace("_", " ")}`,
    detail: plan.steps.length ? `${plan.steps.length} step${plan.steps.length === 1 ? "" : "s"} planned` : "No steps saved yet",
    tone: plan.approvalStatus === "approved" ? "done" : plan.approvalStatus === "changes_requested" || plan.approvalStatus === "rejected" ? "warn" : "active",
  });

  for (const session of sessions) {
    entries.push({
      id: `session-${session.id}`,
      icon: Terminal,
      label: `${session.step ?? session.agentType} agent · ${session.status.replace("_", " ")}`,
      detail: session.command,
      tone: session.status === "completed" || session.status === "stopped" ? "done" : session.status === "failed" ? "warn" : "active",
    });
  }

  if (diff) {
    entries.push({
      id: "diff",
      icon: GitCompare,
      label: "Diff collected",
      detail: `${diff.changedFiles.length} file${diff.changedFiles.length === 1 ? "" : "s"} changed`,
      tone: "done",
    });
  }

  if (test) {
    entries.push({
      id: "test",
      icon: test.status === "passed" ? CheckCircle2 : AlertTriangle,
      label: `Tests ${test.status}`,
      detail: test.command,
      tone: test.status === "passed" ? "done" : "warn",
    });
  }

  entries.push({
    id: "review",
    icon: ShieldCheck,
    label: `Review · ${review.status.replace("_", " ")}`,
    detail: review.reviewerNotes || (review.changedFiles.length ? `${review.changedFiles.length} files reviewed` : "Awaiting review"),
    tone: review.status === "approved" ? "done" : review.status === "rejected" ? "warn" : task.reviewApproved ? "done" : "pending",
  });

  return entries;
}

const EVENT_META: Record<TaskEventType, { icon: LucideIcon; tone: TimelineEntry["tone"]; label: (event: TaskEvent) => string }> = {
  created: { icon: ListTodo, tone: "done", label: () => "Task created" },
  moved: { icon: ArrowRight, tone: "active", label: (e) => `Moved to ${e.to ? statusLabel(e.to) : "next"}` },
  plan_approved: { icon: CheckCircle2, tone: "done", label: () => "Plan approved → Ready" },
  changes_requested: { icon: RotateCcw, tone: "warn", label: () => "Changes requested" },
  agent_started: { icon: Play, tone: "active", label: () => "Agent started (In Progress)" },
  tests_passed: { icon: FlaskConical, tone: "done", label: () => "Tests passed" },
  review_approved: { icon: ShieldCheck, tone: "done", label: () => "Review approved" },
  blocked: { icon: Ban, tone: "warn", label: () => "Blocked" },
  failed: { icon: XCircle, tone: "warn", label: () => "Failed" },
  done: { icon: CheckCircle2, tone: "done", label: () => "Done" },
};

function formatTime(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleString([], { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" });
}

function eventEntries(events: TaskEvent[]): TimelineEntry[] {
  return events.map((event) => {
    const entry = EVENT_META[event.type];
    return {
      id: event.id,
      icon: entry.icon,
      label: entry.label(event),
      detail: [formatTime(event.at), event.note].filter(Boolean).join(" · ") || undefined,
      tone: entry.tone,
    };
  });
}

export function TimelinePanel({ task }: { task: Task | null }) {
  const state = useWorkbenchStore();
  if (!task) return <EmptyState label="No task selected" />;
  // Prefer the recorded workflow history; fall back to a synthesized view.
  const events = taskEventsFor(state, task.id);
  const entries = events.length ? eventEntries(events) : buildTimeline(task, state);
  return (
    <div className="h-full min-h-0 overflow-y-auto">
      <ol className="relative grid gap-4 pl-2">
        {entries.map((entry, index) => {
          const Icon = entry.icon;
          return (
            <li key={entry.id} className="relative flex gap-3">
              <div className="flex flex-col items-center">
                <span className={`grid h-7 w-7 shrink-0 place-items-center rounded-full border border-slate-800 bg-bg-card ${toneClass[entry.tone]}`}>
                  <Icon size={14} />
                </span>
                {index < entries.length - 1 && <span className="mt-1 w-px flex-1 bg-slate-800" />}
              </div>
              <div className="min-w-0 pb-1">
                <p className="text-sm font-medium text-text-main">{entry.label}</p>
                {entry.detail && <p className="mt-0.5 truncate text-xs text-text-muted">{entry.detail}</p>}
              </div>
            </li>
          );
        })}
      </ol>
    </div>
  );
}

export function LogsPanel({ task }: { task: Task | null }) {
  const state = useWorkbenchStore();
  if (!task) return <EmptyState label="No task selected" />;
  const sessions = state.sessions.filter((session) => session.taskId === task.id && session.output.length);
  if (!sessions.length) {
    return (
      <div className="grid h-full place-items-center">
        <div className="flex flex-col items-center gap-2 text-text-muted">
          <ScrollText size={20} />
          <p className="text-sm">No agent logs yet. Start an agent to stream logs here.</p>
        </div>
      </div>
    );
  }
  return (
    <div className="h-full min-h-0 space-y-3 overflow-y-auto">
      {sessions.map((session) => (
        <div key={session.id} className="rounded-lg border border-slate-800 bg-slate-950/70">
          <div className="flex items-center justify-between gap-2 border-b border-slate-800 px-3 py-1.5 text-xs">
            <span className="inline-flex items-center gap-2 font-medium text-text-main">
              <Circle size={8} className={session.status === "running" ? "fill-green-success text-green-success" : "fill-slate-600 text-slate-600"} />
              {session.step ?? session.agentType} · {session.provider}
            </span>
            <span className="text-text-muted">{session.status.replace("_", " ")}</span>
          </div>
          <pre className="max-h-64 overflow-auto whitespace-pre-wrap px-3 py-2 font-mono text-[11px] leading-relaxed text-text-muted">
            {session.output.join("")}
          </pre>
        </div>
      ))}
    </div>
  );
}
