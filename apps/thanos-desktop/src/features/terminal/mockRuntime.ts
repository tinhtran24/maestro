// Phase 5 — Native Terminal Runtime.
// A MOCK terminal runtime. Phase 5 is "terminal UI complete, no actual
// execution required": sessions stream simulated output in real time so the
// full terminal experience (tabs, logs, status, restart/stop, transcript) is
// visible without running any real process. Provider/command come from the
// Phase 4 workflow-step config — nothing is hardcoded or actually executed.

import type { AgentSession, Task, WorkflowStepId } from "../../domain/models";
import { useWorkbenchStore, workflowStepFor } from "../../state/workbenchStore";

export type TerminalStep = "planning" | "coding" | "review" | "testing";

const STEP_META: Record<TerminalStep, { agentType: AgentSession["agentType"]; label: string }> = {
  planning: { agentType: "planner", label: "Planning" },
  coding: { agentType: "coder", label: "Coding" },
  review: { agentType: "reviewer", label: "Review" },
  testing: { agentType: "tester", label: "Tests" },
};

const STREAM_MS = 300;
const timers = new Map<string, ReturnType<typeof setInterval>>();

export function sessionIdFor(taskId: string, step: TerminalStep): string {
  return `sess-${taskId}-${step}`;
}

export function terminalLabel(step: WorkflowStepId | undefined, provider: string): string {
  const label = step && step in STEP_META ? STEP_META[step as TerminalStep].label : (step ?? "session");
  return `${label} · ${provider}`;
}

// Pure: the simulated output for a step. Exported for tests.
export function scriptFor(step: TerminalStep, command: string, taskTitle: string): string[] {
  switch (step) {
    case "planning":
      return [
        "● Thanos planning session — Claude Code",
        `Reading task: ${taskTitle}`,
        "Scanning repository structure…",
        "Identifying impacted modules…",
        "Drafting execution plan:",
        "  1. Analyze requirements and data model",
        "  2. Implement core changes",
        "  3. Add tests and documentation",
        "Estimated risk: low",
        "Plan ready — waiting for user approval.",
      ];
    case "coding":
      return [
        "● Thanos coding session — Codex",
        "Checked out isolated worktree branch",
        "Applying planned changes…",
        "  edit  src/domain/models.ts        (+24 -2)",
        "  edit  src/state/workbenchStore.ts  (+58 -6)",
        "  add   src/features/terminal/…      (+120)",
        "Running formatter…",
        "Changes staged. Ready for review.",
      ];
    case "review":
      return [
        "● Thanos review session — Claude Code",
        "Collecting diff for review…",
        "12 files changed, 240 insertions(+), 18 deletions(-)",
        "Checklist:",
        "  [x] Tests present",
        "  [x] No secrets committed",
        "  [x] Follows project conventions",
        "Review notes drafted.",
      ];
    case "testing":
      return [
        `$ ${command}`,
        "Running test suite…",
        "PASS  src/state/taskMachine.test.ts",
        "PASS  src/features/agents/api/agentCatalog.test.ts",
        "Test Files  2 passed (2)",
        "Tests  15 passed (15)",
      ];
  }
}

function clearTimer(id: string) {
  const timer = timers.get(id);
  if (timer) {
    clearInterval(timer);
    timers.delete(id);
  }
}

// Starts (or restarts) a mock session for a step and begins streaming output.
export function startSession(task: Task, step: TerminalStep): string {
  const store = useWorkbenchStore.getState();
  const config = workflowStepFor(store, step);
  const meta = STEP_META[step];
  const command = [config.command, ...config.args].filter(Boolean).join(" ") || config.command || meta.agentType;
  const id = sessionIdFor(task.id, step);
  clearTimer(id);

  const cwd = config.workingDirectoryMode === "worktree"
    ? task.worktreePath || `.thanos/worktrees/${task.id.toLowerCase()}`
    : ".";

  store.upsertSession({
    id,
    taskId: task.id,
    step,
    agentType: meta.agentType,
    provider: config.provider,
    command,
    cwd,
    status: "starting",
    transcriptPath: `.thanos/logs/sessions/${id}.log`,
    startedAt: new Date().toISOString(),
    output: [
      "\x1b[38;5;244m[simulated — real agent unavailable in this environment]\x1b[0m\r\n",
      `$ ${command}\r\n`,
    ],
  });
  store.setActiveSession(task.id, id);

  const lines = scriptFor(step, command, task.title);
  let index = 0;
  const timer = setInterval(() => {
    const state = useWorkbenchStore.getState();
    const current = state.sessions.find((session) => session.id === id);
    if (!current || current.status === "stopped") {
      clearTimer(id);
      return;
    }
    if (current.status === "starting") state.patchSession(id, { status: "running" });
    if (index < lines.length) {
      state.appendSessionOutput(task.id, id, `${lines[index]}\r\n`);
      index += 1;
      return;
    }
    clearTimer(id);
    state.appendSessionOutput(task.id, id, "[process exited with code 0]\r\n");
    state.patchSession(id, { status: "completed", endedAt: new Date().toISOString() });
  }, STREAM_MS);
  timers.set(id, timer);
  return id;
}

export function stopSession(sessionId: string) {
  clearTimer(sessionId);
  const store = useWorkbenchStore.getState();
  const session = store.sessions.find((item) => item.id === sessionId);
  if (!session || session.status === "completed" || session.status === "stopped") return;
  store.appendSessionOutput(session.taskId, sessionId, "^C\r\n");
  store.appendSessionOutput(session.taskId, sessionId, "[session stopped by user]\r\n");
  store.patchSession(sessionId, { status: "stopped", endedAt: new Date().toISOString() });
}

export function restartSession(sessionId: string) {
  const store = useWorkbenchStore.getState();
  const session = store.sessions.find((item) => item.id === sessionId);
  if (!session || !session.step) return;
  const task = store.tasks.find((item) => item.id === session.taskId);
  if (!task) return;
  startSession(task, session.step as TerminalStep);
}
