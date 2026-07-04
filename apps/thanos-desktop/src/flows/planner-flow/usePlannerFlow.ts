// Phase 6 — Planner Workflow (mock).
// Orchestrates the planning lifecycle for a task: start planning → answer
// clarifying questions → generate an execution plan → approval gate. Gates and
// transitions run through the validated state machine; no coding before approval.

import { useMemo } from "react";
import type { Task } from "../../domain/models";
import { generatePlan, generateQuestions } from "../../features/planner/planGenerator";
import { startSession } from "../../features/terminal/mockRuntime";
import { emptyPlan, planningFor, useWorkbenchStore } from "../../state/workbenchStore";

export type PlannerStage = "start" | "questions" | "review" | "approved";

export function usePlannerFlow(task: Task) {
  const questions = useWorkbenchStore((state) => planningFor(state, task.id));
  // Stable stored plan + fallback outside the selector (avoids an infinite loop).
  const savedPlan = useWorkbenchStore((state) => state.plans.find((plan) => plan.taskId === task.id));
  const plan = useMemo(() => savedPlan ?? emptyPlan(task.id), [savedPlan, task.id]);
  const setPlanningQuestions = useWorkbenchStore((state) => state.setPlanningQuestions);
  const answerPlanningQuestion = useWorkbenchStore((state) => state.answerPlanningQuestion);
  const persistPlan = useWorkbenchStore((state) => state.persistPlan);
  const advanceTask = useWorkbenchStore((state) => state.advanceTask);
  const approvePlan = useWorkbenchStore((state) => state.approvePlan);
  const requestChanges = useWorkbenchStore((state) => state.requestChanges);
  const setBottomTab = useWorkbenchStore((state) => state.setBottomTab);

  const answered = questions.length > 0 && questions.every((question) => question.answer.trim().length > 0);

  const stage: PlannerStage = (() => {
    if (["ready", "running", "in_review", "done"].includes(task.status)) return "approved";
    if (task.status === "waiting_approval") return "review";
    if (questions.length > 0) return "questions";
    return "start";
  })();

  function startPlanning() {
    if (task.status === "backlog") advanceTask(task.id, "planning");
    if (planningFor(useWorkbenchStore.getState(), task.id).length === 0) {
      setPlanningQuestions(task.id, generateQuestions(task));
    }
    setBottomTab("terminal");
    startSession(task, "planning");
  }

  function answer(questionId: string, value: string) {
    answerPlanningQuestion(task.id, questionId, value);
  }

  function submitPlan() {
    if (!answered) return;
    persistPlan(generatePlan(task, questions));
    advanceTask(task.id, "waiting_approval");
  }

  function approve() {
    approvePlan(task.id);
  }

  function reject() {
    // Send back to planning; keep answers so the user can revise and regenerate.
    persistPlan({ ...plan, approvalStatus: "changes_requested" });
    requestChanges(task.id);
  }

  return { stage, questions, plan, answered, startPlanning, answer, submitPlan, approve, reject };
}
