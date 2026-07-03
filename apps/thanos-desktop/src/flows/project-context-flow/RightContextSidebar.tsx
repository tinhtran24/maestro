import { Check, ChevronRight, GitCompare, Inbox, Play, RotateCcw, Sparkles, Terminal } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { Task } from "../../domain/models";
import { Card } from "../../shared/ui/Card";
import { PriorityBadge } from "../../shared/ui/PriorityBadge";
import { StatusBadge } from "../../shared/ui/StatusBadge";
import { activeSkillsFor, currentWorkflowStep, selectedTask, useWorkbenchStore, workflowStepFor } from "../../state/workbenchStore";
import { TaskActivityCard } from "../task-workbench-flow/TaskActivityCard";
import { WorkflowControls } from "../task-workbench-flow/WorkflowControls";
import { WorkflowTimelineCard } from "../task-workbench-flow/WorkflowTimelineCard";
import { ActiveWorkflowCard } from "./ActiveWorkflowCard";
import { ProjectInfoCard } from "./ProjectInfoCard";
import { QuickActionsCard } from "./QuickActionsCard";

export function RightContextSidebar() {
  const state = useWorkbenchStore();
  const task = selectedTask(state);
  const toggle = state.toggleRightCollapsed;

  return (
    <aside className="grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] border-l border-slate-800 bg-slate-900/40">
      <header className="flex items-center justify-between border-b border-slate-800 bg-slate-950/50 px-4 py-3 backdrop-blur">
        <h2 className="inline-flex items-center gap-2 text-sm font-semibold">
          <ChevronRight size={15} className="text-text-muted" />
          Task Details
        </h2>
        <button onClick={toggle} className="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs text-text-muted hover:bg-slate-800 hover:text-text-main">
          Collapse
        </button>
      </header>
      <div className="min-h-0 space-y-4 overflow-y-auto p-4">
        {task ? <TaskContext task={task} /> : <EmptyContext />}
      </div>
    </aside>
  );
}

function EmptyContext() {
  const project = useWorkbenchStore((state) => state.project);
  return (
    <>
      <div className="grid place-items-center rounded-xl border border-dashed border-slate-700 bg-slate-950/40 px-4 py-10 text-center">
        <span className="grid h-12 w-12 place-items-center rounded-full border border-slate-800 bg-slate-900 text-text-muted"><Inbox size={22} /></span>
        <p className="mt-3 text-sm font-medium text-text-main">No task selected</p>
        <p className="mt-1 text-xs text-text-muted">Select a task from the board to view details</p>
      </div>
      <QuickActionsCard />
      <ActiveWorkflowCard task={null} />
      <ProjectInfoCard project={project} />
    </>
  );
}

function TaskContext({ task }: { task: Task }) {
  const state = useWorkbenchStore();
  const step = workflowStepFor(state, currentWorkflowStep(task));
  const activeSkills = activeSkillsFor(task, state);
  const focusTerminal = () => { state.setActiveView("workbench"); state.setBottomTab("terminal"); };

  const actions: Array<{ icon: LucideIcon; label: string; onClick: () => void; primary?: boolean }> = [
    { icon: Check, label: "Approve Plan", onClick: () => state.approvePlan(task.id), primary: true },
    { icon: RotateCcw, label: "Request Changes", onClick: () => state.requestChanges(task.id) },
    { icon: Play, label: "Start Agent", onClick: focusTerminal },
    { icon: Terminal, label: "Open Terminal", onClick: focusTerminal },
    { icon: Play, label: "Run Tests", onClick: () => state.runTests(task.id) },
    { icon: GitCompare, label: "Open Git Diff", onClick: focusTerminal },
  ];

  return (
    <>
      <Card>
        <div className="flex items-center justify-between gap-2">
          <span className="text-xs font-medium text-text-muted">{task.id}</span>
          <PriorityBadge priority={task.priority} />
        </div>
        <h3 className="mt-2 text-base font-semibold text-text-main">{task.title}</h3>
        <p className="mt-1 text-sm text-text-muted">{task.description}</p>
        <div className="mt-3 flex flex-wrap items-center gap-2">
          <StatusBadge status={task.status} />
          <span className="inline-flex items-center gap-1 rounded-lg border border-slate-700 bg-slate-800/60 px-2 py-1 text-xs text-text-muted">{step.label}</span>
        </div>
      </Card>

      <div className="grid grid-cols-2 gap-2">
        {actions.map(({ icon: Icon, label, onClick, primary }) => (
          <button key={label} onClick={onClick} className={`inline-flex items-center justify-center gap-2 rounded-lg border px-2 py-2 text-xs font-medium transition ${primary ? "border-purple-primary bg-purple-primary text-white hover:bg-purple-hover" : "border-slate-800 bg-slate-900/70 text-text-main hover:border-slate-700"}`}>
            <Icon size={14} /> {label}
          </button>
        ))}
      </div>

      <WorkflowControls task={task} />
      <WorkflowTimelineCard task={task} />
      <ActiveWorkflowCard task={task} />
      <TaskActivityCard task={task} />

      <Card title="Active Skills" icon={Sparkles} action={<span className="text-xs text-text-muted">{activeSkills.length}</span>}>
        {activeSkills.length ? (
          <ul className="grid gap-2">
            {activeSkills.map(({ skill, run }) => (
              <li key={run.id} className="rounded-lg border border-slate-800 bg-bg-card p-3">
                <div className="flex items-start justify-between gap-2">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">{skill.name}</p>
                    <p className="mt-0.5 truncate text-xs text-text-muted">{skill.agents.join(", ")}</p>
                  </div>
                  <span className="shrink-0 rounded-md border border-blue-info/30 bg-blue-info/10 px-2 py-0.5 text-[11px] text-blue-info">{run.status.replace("_", " ")}</span>
                </div>
              </li>
            ))}
          </ul>
        ) : (
          <p className="rounded-lg border border-slate-800 bg-bg-card p-3 text-xs text-text-muted">No active skills matched for this task.</p>
        )}
      </Card>

      <ProjectInfoCard project={state.project} />
    </>
  );
}
