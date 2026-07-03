// Phase 3 — Workflow Engine.
// Validated task state machine. This is the single source of truth for which
// transitions are legal and which gates guard them. No invalid transition may
// be produced: every mutation must route through `transitionTask`.

import type { Task, TaskEventType, TaskStatus } from "../domain/models";

// Allowed transitions per status. Anything not listed here is invalid.
const transitions: Record<TaskStatus, TaskStatus[]> = {
  backlog: ["planning"],
  planning: ["waiting_approval", "blocked", "failed"],
  waiting_approval: ["planning", "ready", "blocked"],
  ready: ["running", "planning", "blocked"],
  running: ["in_review", "blocked", "failed"],
  in_review: ["ready", "running", "blocked", "done", "failed", "waiting_user"],
  waiting_user: ["planning", "ready", "running", "in_review", "blocked", "failed"],
  blocked: ["planning", "ready", "running", "failed"],
  failed: ["planning", "ready"],
  done: [],
};

// Canonical happy-path order, used for stepper/progress visualisation.
export const MAIN_FLOW: TaskStatus[] = [
  "backlog",
  "planning",
  "waiting_approval",
  "ready",
  "running",
  "in_review",
  "done",
];

const STATUS_LABELS: Record<TaskStatus, string> = {
  backlog: "Backlog",
  planning: "Planning",
  waiting_approval: "Waiting Approval",
  ready: "Ready",
  running: "In Progress",
  in_review: "In Review",
  waiting_user: "Waiting User",
  blocked: "Blocked",
  failed: "Failed",
  done: "Done",
};

export function statusLabel(status: TaskStatus): string {
  return STATUS_LABELS[status];
}

export type TransitionResult = { ok: true; task: Task } | { ok: false; reason: string };

export type AvailableTransition = { to: TaskStatus; allowed: boolean; reason: string | null };

// Returns null when the transition is allowed, otherwise a human-readable reason.
export function canTransition(task: Task, to: TaskStatus): string | null {
  if (!transitions[task.status].includes(to)) {
    return `Invalid transition ${statusLabel(task.status)} → ${statusLabel(to)}.`;
  }
  // Gate: plan approval is required before a task becomes ready.
  if (to === "ready" && task.status === "waiting_approval" && !task.planApproved) {
    return "Plan approval is required before the task can become ready.";
  }
  // Gate: an isolated worktree + branch is required before execution (mock-assigned in Phase 3).
  if (to === "running" && (!task.worktreePath || !task.branchName)) {
    return "An isolated worktree and branch are required before agent execution.";
  }
  // Gate: done requires an approved review and passing tests.
  if (to === "done" && !task.reviewApproved) {
    return "Review approval is required before the task can be done.";
  }
  if (to === "done" && !task.testsPassed) {
    return "Passing tests are required before the task can be done.";
  }
  return null;
}

// All transition targets for a task's current status, each annotated with
// whether it is currently allowed and why not.
export function availableTransitions(task: Task): AvailableTransition[] {
  return transitions[task.status].map((to) => {
    const reason = canTransition(task, to);
    return { to, allowed: reason === null, reason };
  });
}

// Maps a target status onto the lifecycle event type it records.
export function transitionEventType(to: TaskStatus): TaskEventType {
  switch (to) {
    case "ready":
      return "plan_approved";
    case "running":
      return "agent_started";
    case "in_review":
      return "review_approved";
    case "blocked":
      return "blocked";
    case "failed":
      return "failed";
    case "done":
      return "done";
    default:
      return "moved";
  }
}

export function transitionTask(task: Task, to: TaskStatus): TransitionResult {
  const reason = canTransition(task, to);
  if (reason) {
    return { ok: false, reason };
  }
  return { ok: true, task: { ...task, status: to, updatedAt: new Date().toISOString() } };
}
