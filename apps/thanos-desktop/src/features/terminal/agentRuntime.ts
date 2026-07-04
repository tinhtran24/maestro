// Real agent runtime. Drives the Tauri PTY backend (start_agent_session) so the
// terminal runs actual `claude`/`codex` processes, streaming real stdout/stderr
// into the session output and routing keystrokes via write_terminal (wired in
// XtermPanel). Mirrors the mockRuntime API so call sites barely change.
//
// When the backend is unavailable (browser dev, no Tauri) it falls back to the
// clearly-labeled mock simulation; when the backend is present but the start
// fails (missing CLI, no worktree, plan not approved), the reason is shown in the
// terminal instead of being silently swallowed.

import type { AgentSession, Task } from "../../domain/models";
import { NativeBackend } from "../../services/nativeBackend";
import { useWorkbenchStore } from "../../state/workbenchStore";
import { sessionIdFor, startSession as startMockSession, stopSession as stopMockSession, restartSession as restartMockSession, type TerminalStep } from "./mockRuntime";

const backend = new NativeBackend();

const STEP_AGENT: Record<TerminalStep, AgentSession["agentType"]> = {
  planning: "planner",
  coding: "coder",
  review: "reviewer",
  testing: "tester",
};

// Exported for tests.
export function agentTypeFor(step: TerminalStep): AgentSession["agentType"] {
  return STEP_AGENT[step];
}

export function hasBackend(): boolean {
  return typeof window !== "undefined" && "__TAURI_INTERNALS__" in window;
}

// Mock sessions are keyed `sess-<task>-<step>`; real sessions are keyed
// `<task>-<agentType>-<epoch>`. Used to route stop/restart correctly.
function isSimulated(session: AgentSession): boolean {
  return session.id.startsWith("sess-");
}

let listenersReady = false;
// Subscribe once to the backend's output/exit streams and fan them into the
// store. Idempotent; no-op without a backend.
export function ensureListeners() {
  if (listenersReady || !hasBackend()) return;
  listenersReady = true;
  void backend.onAgentOutput((taskId, sessionId, data) => {
    useWorkbenchStore.getState().appendSessionOutput(taskId, sessionId, data);
  });
  void backend.onAgentExit((taskId, sessionId, code) => {
    const store = useWorkbenchStore.getState();
    store.appendSessionOutput(taskId, sessionId, `\r\n[process exited with code ${code}]\r\n`);
    store.patchSession(sessionId, { status: code === 0 ? "completed" : "failed", endedAt: new Date().toISOString() });
  });
}

function showFailure(task: Task, step: TerminalStep, message: string): string {
  const store = useWorkbenchStore.getState();
  const id = sessionIdFor(task.id, step);
  store.upsertSession({
    id,
    taskId: task.id,
    step,
    agentType: STEP_AGENT[step],
    provider: "agent",
    command: step,
    status: "failed",
    output: [`\x1b[31m⚠ Could not start the ${step} agent:\x1b[0m ${message}\r\n`],
  });
  store.setActiveSession(task.id, id);
  return id;
}

// Starts a real agent session for a workflow step. Returns the session id.
export async function startAgentStep(task: Task, step: TerminalStep): Promise<string> {
  ensureListeners();
  const store = useWorkbenchStore.getState();

  if (step === "coding" && !task.planApproved) {
    return showFailure(task, step, "Coding is blocked until the execution plan is approved.");
  }

  // No desktop backend (browser dev): fall back to the labeled simulation.
  if (!hasBackend()) {
    return startMockSession(task, step);
  }

  const agentType = STEP_AGENT[step];
  // The coder must run inside an isolated git worktree.
  let runnable = task;
  if (agentType === "coder") {
    const prepared = await backend.prepareWorktree(task);
    if (prepared) runnable = { ...task, branchName: prepared.branchName, worktreePath: prepared.worktreePath };
  }

  const result = await backend.startAgentRoleResult(runnable, agentType);
  if ("session" in result) {
    store.upsertSession({ ...result.session, step });
    store.setActiveSession(task.id, result.session.id);
    return result.session.id;
  }
  return showFailure(task, step, result.error);
}

export function stopAgentStep(session: AgentSession | null) {
  if (!session) return;
  if (isSimulated(session)) {
    stopMockSession(session.id);
    return;
  }
  void backend.stopAgent();
  useWorkbenchStore.getState().patchSession(session.id, { status: "stopped", endedAt: new Date().toISOString() });
}

export function restartAgentStep(session: AgentSession | null) {
  if (!session?.step) return;
  const task = useWorkbenchStore.getState().tasks.find((item) => item.id === session.taskId);
  if (!task) return;
  if (isSimulated(session)) {
    restartMockSession(session.id);
    return;
  }
  void backend.stopAgent();
  void startAgentStep(task, session.step as TerminalStep);
}
