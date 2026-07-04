// Phase 7 — Coding Workflow (mock).
// Orchestrates the coding lifecycle for an approved task: create a mock worktree
// → open the coding (Codex) terminal → stream coding output → send to review.
// Transitions run through the validated state machine; the mock worktree is
// assigned on the `running` transition and coding cannot be sent to review until
// the coding session completes. No real Git and no real agent execution.

import { useMemo } from "react";
import type { Task } from "../../domain/models";
import { generateChangeset } from "../../features/coder/changesetGenerator";
import { sessionIdFor, startSession } from "../../features/terminal/mockRuntime";
import { emptyPlan, useWorkbenchStore } from "../../state/workbenchStore";

export type CoderStage = "blocked" | "ready" | "coding" | "review" | "done";

export function useCoderFlow(task: Task) {
  // Subscribe to the stored plan (stable ref) and fall back outside the selector —
  // returning a fresh default inside the selector causes an infinite render loop.
  const savedPlan = useWorkbenchStore((state) => state.plans.find((plan) => plan.taskId === task.id));
  const plan = useMemo(() => savedPlan ?? emptyPlan(task.id), [savedPlan, task.id]);
  const diff = useWorkbenchStore((state) => state.diffs[task.id]);
  const session = useWorkbenchStore((state) =>
    state.sessions.find((item) => item.id === sessionIdFor(task.id, "coding")),
  );
  const advanceTask = useWorkbenchStore((state) => state.advanceTask);
  const persistDiff = useWorkbenchStore((state) => state.persistDiff);
  const setBottomTab = useWorkbenchStore((state) => state.setBottomTab);

  const codingComplete = session?.status === "completed";

  const stage: CoderStage = (() => {
    if (["in_review", "done"].includes(task.status)) return "done";
    if (!task.planApproved) return "blocked";
    if (task.status === "ready") return "ready";
    // running
    return codingComplete ? "review" : "coding";
  })();

  // Approved → Create Worktree → Open Codex. The `running` transition mock-assigns
  // the worktree/branch and records the `agent_started` timeline event; we then
  // read the freshly-assigned task back before starting the coding terminal.
  function startCoding() {
    if (task.status !== "ready" || !task.planApproved) return;
    advanceTask(task.id, "running");
    const runnable = useWorkbenchStore.getState().tasks.find((item) => item.id === task.id) ?? task;
    persistDiff(generateChangeset(runnable, plan));
    setBottomTab("terminal");
    startSession(runnable, "coding");
  }

  // Exit criteria: task enters Review. Gated on the coding session completing.
  function sendToReview() {
    if (!codingComplete) return;
    advanceTask(task.id, "in_review");
  }

  return { stage, plan, diff, session, codingComplete, startCoding, sendToReview };
}
