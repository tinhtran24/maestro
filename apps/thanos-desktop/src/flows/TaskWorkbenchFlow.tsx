import { ArrowRight, Check, Clock3, GitCompare, Inbox, MessageSquare, Play, RotateCcw, ScrollText, Send, Square, Terminal, Trash2 } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { BottomTab } from "../domain/models";
import { FlowTabs } from "../shared/ui/FlowTabs";
import { StatusBadge } from "../shared/ui/StatusBadge";
import { currentWorkflowStep, selectedTask, planFor, reviewFor, sessionFor, useWorkbenchStore, workflowStepFor } from "../state/workbenchStore";
import { useAgentSessionFlow, XtermPanel } from "./AgentSessionFlow";
import { PlannerFlow, isPlanningPhase } from "./planner-flow/PlannerFlow";
import { CoderFlow, isCodingPhase } from "./coder-flow/CoderFlow";
import { AgentTerminalFlow } from "./agent-terminal-flow/AgentTerminalFlow";
import { LogsPanel, TimelinePanel } from "./task-workbench-flow/BottomPanels";
import { StartAgentButton } from "./task-workbench-flow/StartAgentButton";
import { StepApprovalGate } from "./task-workbench-flow/StepApprovalGate";
import { TaskStepPanel } from "./task-workbench-flow/TaskStepPanel";

// Bottom panel tabs — live runtime streams for the selected task.
const bottomTabs: Array<{ id: BottomTab; label: string; icon: LucideIcon }> = [
  { id: "terminal", label: "Terminal", icon: Terminal },
  { id: "timeline", label: "Timeline", icon: Clock3 },
  { id: "logs", label: "Logs", icon: ScrollText },
  { id: "chat", label: "Chat", icon: MessageSquare },
];

export function TaskWorkbenchFlow() {
  return <TaskWorkbenchMain />;
}

export function TaskWorkbenchMain() {
  const state = useWorkbenchStore();
  const task = selectedTask(state);
  const agentFlow = useAgentSessionFlow();
  if (!task) {
    return (
      <section className="grid h-full min-h-0 place-items-center border-t border-slate-800 bg-bg-app p-6">
        <div className="relative flex w-full max-w-2xl items-center justify-center">
          <div className="grid place-items-center rounded-2xl border border-dashed border-slate-700 bg-slate-950/40 px-10 py-12 text-center">
            <span className="grid h-14 w-14 place-items-center rounded-full border border-slate-800 bg-slate-900 text-text-muted"><Inbox size={26} /></span>
            <p className="mt-4 text-base font-semibold text-text-main">No task selected</p>
            <p className="mt-1 text-sm text-text-muted">Select a task from the board to see details.</p>
          </div>
          <ArrowRight size={80} className="pointer-events-none absolute -right-2 hidden text-slate-800 xl:block" strokeWidth={1} />
        </div>
      </section>
    );
  }
  const plan = planFor(task, state);
  const review = reviewFor(task, state);
  const session = sessionFor(task, state);
  const stepId = currentWorkflowStep(task);
  const step = workflowStepFor(state, stepId);
  const startStep = async (id: typeof stepId) => {
    state.setBottomTab("terminal");
    if (id === "planning") await agentFlow.start(task, "planner");
    else if (id === "coding" && task.planApproved) await agentFlow.start(task, "coder");
    else if (id === "review") await agentFlow.start(task, "reviewer");
    else if (id === "testing") await runTests();
  };
  const approvePlan = async () => {
    const saved = await agentFlow.savePlan(plan);
    if (saved) state.persistPlan(saved);
    const approved = await agentFlow.approvePlan(task.id);
    if (approved) state.persistPlan(approved);
    state.approvePlan(task.id);
  };
  const collectDiff = async () => {
    const diff = await agentFlow.collectDiff(task);
    if (diff) state.persistDiff(diff);
  };
  const runTests = async () => {
    const test = await agentFlow.runTests(task);
    if (test) state.persistTestRun(test);
    else state.runTests(task.id);
  };
  const approveReview = async () => {
    const diff = state.diffs[task.id];
    const test = state.testRuns[task.id];
    const draft = {
      id: `review-${task.id}`,
      taskId: task.id,
      diffSummary: diff?.summary || review.diffSummary,
      changedFiles: diff?.changedFiles.map((file) => file.path) || review.changedFiles,
      testResults: test ? [{ command: test.command, status: test.status, output: test.stdout || test.stderr }] : review.testResults,
      reviewerNotes: "User approved review from workbench.",
      status: "pending" as const,
    };
    const saved = await agentFlow.saveReview(draft);
    if (saved) state.persistReview(saved);
    const approved = await agentFlow.approveReview(task.id);
    if (approved) state.persistReview(approved);
    const memory = await agentFlow.searchMemory(task.title.split(" ")[0] || task.id);
    if (memory.length) state.persistMemory(memory);
  };

  return (
    <section className="grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] border-t border-slate-800 bg-bg-app">
      <header className="sticky top-0 z-10 flex flex-wrap items-center justify-between gap-3 border-b border-slate-800 bg-slate-950/80 p-4 backdrop-blur">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <h2 className="text-lg font-semibold">{task.title}</h2>
            <StatusBadge status={task.status} />
          </div>
          <p className="mt-1 text-sm text-text-muted">{task.id} · {task.description}</p>
        </div>
        <div className="flex min-w-0 items-center gap-2 overflow-x-auto pb-1">
          <StartAgentButton label="Start Planning Agent" step="planning" onStart={startStep} />
          <Action icon={Check} label="Approve" onClick={approvePlan} primary />
          <Action icon={RotateCcw} label="Changes" onClick={() => state.requestChanges(task.id)} />
          <Action icon={GitCompare} label="Diff" onClick={collectDiff} />
          <Action icon={Play} label="Run Tests" onClick={runTests} />
          <Action icon={Check} label="Review" onClick={approveReview} />
          <StartAgentButton label="Start Coding Agent" step="coding" onStart={startStep} disabled={!task.planApproved} />
          <Action icon={Trash2} label="Remove" onClick={() => state.removeTask(task.id)} danger />
          <Action icon={Square} label="Stop" onClick={() => agentFlow.stop()} danger />
        </div>
      </header>
      <div className="min-h-0 overflow-y-auto p-4">
        {isPlanningPhase(task) && (
          <div className="mb-3">
            <PlannerFlow task={task} />
          </div>
        )}
        {isCodingPhase(task) && (
          <div className="mb-3">
            <CoderFlow task={task} />
          </div>
        )}
        <div className="grid grid-cols-1 gap-3 xl:grid-cols-2 2xl:grid-cols-3">
          <Panel title="Current Step" meta={step.label}>
            <div className="grid gap-3">
              <TaskStepPanel task={task} step={step} session={session} onOpenTerminal={() => state.setBottomTab("terminal")} />
              <StepApprovalGate task={task} step={step} />
            </div>
          </Panel>
          <Panel title="Execution Plan" meta={plan.approvalStatus}>
            <ol className="grid gap-2 text-sm">
              {plan.steps.map((step) => <li key={step.id}><span className="font-medium">{step.title}</span><p className="text-xs text-text-muted">{step.description}</p></li>)}
            </ol>
          </Panel>
          <Panel title="Files" meta="planned">
            <ul className="grid gap-2 text-sm text-text-muted">{plan.filesToTouch.map((file) => <li key={file}>{file}</li>)}</ul>
          </Panel>
          <Panel title="Git Changes" meta="isolated">
            <pre className="max-h-40 overflow-auto whitespace-pre-wrap text-xs text-text-muted">{state.diffs[task.id]?.summary || review.diffSummary || "Collect diff after coder changes."}</pre>
            <ul className="mt-3 grid gap-2 text-sm">{(state.diffs[task.id]?.changedFiles.map((file) => file.path) || review.changedFiles).map((file) => <li key={file}><span className="text-green-success">•</span> {file}</li>)}</ul>
          </Panel>
          <Panel title="Terminal" meta={session.status}>
            <XtermPanel output={session.output} />
          </Panel>
          <Panel title="Browser" meta="preview">
            <div className="rounded-xl bg-slate-950/70 p-4 text-sm"><strong>Shopping Cart</strong><p className="mt-2 text-text-muted">Preview attaches after runtime starts.</p></div>
          </Panel>
          <Panel title="Tests" meta={task.testsPassed ? "passed" : "pending"}>
            <p className="text-sm text-text-muted">{state.testRuns[task.id]?.status || "Run tests from the workbench when implementation is ready."}</p>
            <pre className="mt-2 max-h-32 overflow-auto whitespace-pre-wrap text-xs text-text-muted">{state.testRuns[task.id]?.stdout || state.testRuns[task.id]?.stderr || ""}</pre>
          </Panel>
        </div>
      </div>
    </section>
  );
}

export function TaskBottomPanel() {
  const state = useWorkbenchStore();
  const task = selectedTask(state);
  return (
    <section className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)] border-t border-slate-800 bg-slate-950/70">
      <FlowTabs tabs={bottomTabs} active={state.bottomTab} onChange={state.setBottomTab} />
      <div className="min-h-0 overflow-hidden p-3 text-sm text-text-muted">
        {!task ? (
          <BottomEmptyState />
        ) : (
          <>
            {state.bottomTab === "terminal" && <AgentTerminalFlow task={task} />}
            {state.bottomTab === "timeline" && <TimelinePanel task={task} />}
            {state.bottomTab === "logs" && <LogsPanel task={task} />}
            {state.bottomTab === "chat" && <ChatPanel />}
          </>
        )}
      </div>
    </section>
  );
}

function BottomEmptyState() {
  return (
    <div className="grid h-full place-items-center text-center">
      <div className="flex flex-col items-center gap-2 text-text-muted">
        <Inbox size={22} />
        <p className="text-sm font-medium text-text-main">No task selected</p>
        <p className="text-xs">Logs and terminal output will appear after an agent starts.</p>
      </div>
    </div>
  );
}

function ChatPanel() {
  return (
    <div className="grid h-full min-h-0 grid-rows-[minmax(0,1fr)_auto] gap-2">
      <div className="grid place-items-center rounded-lg border border-slate-800 bg-bg-card text-center text-xs text-text-muted">
        Ask the agent about this task. Messages appear here.
      </div>
      <form onSubmit={(event) => event.preventDefault()} className="flex items-center gap-2">
        <input placeholder="Message the agent…" className="h-9 min-w-0 flex-1 rounded-lg border border-slate-800 bg-slate-950 px-3 text-sm text-text-main placeholder:text-text-muted" />
        <button type="submit" className="inline-flex items-center gap-2 rounded-lg bg-purple-primary px-3 py-2 text-sm font-medium text-white hover:bg-purple-hover">
          <Send size={15} /> Send
        </button>
      </form>
    </div>
  );
}

function Panel({ title, meta, children }: { title: string; meta: string; children: React.ReactNode }) {
  return (
    <article className="min-h-48 rounded-xl border border-slate-800 bg-slate-900/80 p-4 shadow-lg shadow-black/20">
      <header className="mb-3 flex items-center justify-between">
        <h3 className="text-base font-semibold">{title}</h3>
        <span className="text-xs text-text-muted">{meta}</span>
      </header>
      {children}
    </article>
  );
}

function Action({ icon: Icon, label, onClick, primary, danger }: { icon: LucideIcon; label: string; onClick: () => void; primary?: boolean; danger?: boolean }) {
  return (
    <button
      onClick={onClick}
      className={`inline-flex items-center gap-2 rounded-lg border px-3 py-2 text-sm ${
        primary ? "border-purple-primary bg-purple-primary text-white hover:bg-purple-hover" : danger ? "border-red-danger/40 text-red-danger" : "border-slate-800 bg-slate-900/80 text-text-main"
      }`}
    >
      <Icon size={16} />
      {label}
    </button>
  );
}
