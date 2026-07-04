import { useState } from "react";
import type { AgentSession, Task } from "../../domain/models";
import type { TerminalStep } from "../../features/terminal/mockRuntime";
import { restartAgentStep, startAgentStep, stopAgentStep } from "../../features/terminal/agentRuntime";
import { useWorkbenchStore } from "../../state/workbenchStore";

// Phase 5 — Native Terminal Runtime (mock). Owns terminal session interaction
// for the selected task: multiple tabs, active selection, start/stop/restart,
// pinning, close, and transcript viewing.
export function useAgentTerminalFlow(task: Task | null) {
  const allSessions = useWorkbenchStore((state) => state.sessions);
  const pinnedSessions = useWorkbenchStore((state) => state.pinnedSessions);
  const activeByTask = useWorkbenchStore((state) => state.activeSessionByTask);
  const setActiveSession = useWorkbenchStore((state) => state.setActiveSession);
  const togglePinnedSession = useWorkbenchStore((state) => state.togglePinnedSession);
  const removeSession = useWorkbenchStore((state) => state.removeSession);
  const [transcriptId, setTranscriptId] = useState("");

  const taskSessions = task ? allSessions.filter((session) => session.taskId === task.id) : [];
  // Pinned tabs first, preserving insertion order otherwise.
  const sessions = taskSessions
    .slice()
    .sort((a, b) => Number(pinnedSessions.includes(b.id)) - Number(pinnedSessions.includes(a.id)));

  const activeId = task ? activeByTask[task.id] : "";
  const active = sessions.find((session) => session.id === activeId) ?? sessions[0] ?? null;
  const transcript = transcriptId ? taskSessions.find((session) => session.id === transcriptId) ?? null : null;

  function start(step: TerminalStep) {
    if (!task) return;
    if (step === "coding" && !task.planApproved) return;
    void startAgentStep(task, step);
  }

  return {
    task,
    sessions,
    active,
    pinnedSessions,
    transcript,
    codingDisabled: !task?.planApproved,
    select: (id: string) => task && setActiveSession(task.id, id),
    start,
    stop: (session: AgentSession | null) => stopAgentStep(session),
    restart: (session: AgentSession | null) => restartAgentStep(session),
    pin: (id: string) => togglePinnedSession(id),
    close: (id: string) => removeSession(id),
    openTranscript: (id: string) => setTranscriptId(id),
    closeTranscript: () => setTranscriptId(""),
    isPinned: (id: string) => pinnedSessions.includes(id),
  };
}
