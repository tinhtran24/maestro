// Phase 8 — Review (mock).
// Orchestrates the review gate before Done: show the diff, changed files,
// checklist, and tests, then Approve / Reject / Request Changes. A task cannot be
// finished (moved to Done) until the review is approved and tests pass — the
// state machine enforces the same gate. On approval, a memory node is synthesized
// (Phase 9 — "memory updates after approval"). No real diff parser or merge.

import { useMemo } from "react";
import type { Task } from "../../domain/models";
import { buildReviewChecklist, canFinish, checklistComplete, mockTestRun } from "../../features/reviewer/reviewChecklist";
import { memoryFromReview } from "../../features/memory/memorySearch";
import { emptyReview, planFor, reviewFor, useWorkbenchStore, workflowStepFor } from "../../state/workbenchStore";

export function useReviewFlow(task: Task) {
  // Stable stored review + fallback outside the selector (avoids an infinite loop).
  const savedReview = useWorkbenchStore((state) => state.reviews.find((review) => review.taskId === task.id));
  const review = useMemo(() => savedReview ?? emptyReview(task.id), [savedReview, task.id]);
  const diff = useWorkbenchStore((state) => state.diffs[task.id]);
  const test = useWorkbenchStore((state) => state.testRuns[task.id]);
  const persistTestRun = useWorkbenchStore((state) => state.persistTestRun);
  const runTests = useWorkbenchStore((state) => state.runTests);
  const approveReview = useWorkbenchStore((state) => state.approveReview);
  const rejectReview = useWorkbenchStore((state) => state.rejectReview);
  const requestReviewChanges = useWorkbenchStore((state) => state.requestReviewChanges);
  const approveMerge = useWorkbenchStore((state) => state.approveMerge);
  const appendMemory = useWorkbenchStore((state) => state.appendMemory);

  const checklist = buildReviewChecklist(task, diff, test);
  const complete = checklistComplete(checklist);
  const finishable = canFinish(task);

  function run() {
    const command = workflowStepFor(useWorkbenchStore.getState(), "testing").command || "npm test";
    persistTestRun(mockTestRun(task, command));
    runTests(task.id);
  }

  function approve() {
    approveReview(task.id);
    // Memory updates after approval — synthesize a decision node from the plan.
    const state = useWorkbenchStore.getState();
    const plan = planFor(task, state);
    appendMemory(memoryFromReview(task, plan, state.memoryNodes, new Date().toISOString(), reviewFor(task, state)));
  }

  function reject() {
    rejectReview(task.id);
  }

  function requestChanges() {
    requestReviewChanges(task.id);
  }

  function finish() {
    if (!finishable) return;
    approveMerge(task.id);
  }

  return { review, diff, test, checklist, complete, finishable, approved: Boolean(task.reviewApproved), run, approve, reject, requestChanges, finish };
}
