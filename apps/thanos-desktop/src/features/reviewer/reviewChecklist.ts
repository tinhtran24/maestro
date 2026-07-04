// Phase 8 — Review.
// Pure, mock review helpers. Phase 8 is UI-first: the reviewer does not run a real
// agent or parse Git. It derives a deterministic review checklist from the task
// and its collected artifacts (diff, test run) and synthesizes a mock test run.
// No execution, no diff parser, no merge.

import type { GitDiff, Task, TestRun } from "../../domain/models";

export type ChecklistItem = { id: string; label: string; done: boolean };

// The review gate items. `diff` and `test` come from the coder/testing artifacts
// already collected on the task; the last two are mock static checks.
export function buildReviewChecklist(task: Task, diff?: GitDiff, test?: TestRun): ChecklistItem[] {
  return [
    { id: "diff", label: "Changes collected for review", done: Boolean(diff && diff.changedFiles.length) },
    { id: "tests", label: "Tests pass", done: task.testsPassed && test?.status !== "failed" },
    { id: "plan", label: "Implementation matches the approved plan", done: Boolean(task.planApproved) },
    { id: "secrets", label: "No secrets committed", done: true },
    { id: "conventions", label: "Follows project conventions", done: true },
  ];
}

export function checklistComplete(items: ChecklistItem[]): boolean {
  return items.every((item) => item.done);
}

// A task may only be finished (merged to Done) once the review is approved and
// tests pass — this mirrors the state-machine gate and drives the Finish button.
export function canFinish(task: Task): boolean {
  return Boolean(task.reviewApproved) && task.testsPassed;
}

// Synthesizes a passing mock test run for the review panel. Deterministic — no
// real test execution.
export function mockTestRun(task: Task, command: string): TestRun {
  return {
    taskId: task.id,
    command,
    status: "passed",
    stdout: [
      `$ ${command}`,
      "PASS  src/state/taskMachine.test.ts",
      "PASS  src/features/reviewer/reviewChecklist.test.ts",
      "Test Files  passed",
      "Tests  passed",
    ].join("\n"),
    stderr: "",
    code: 0,
  };
}
