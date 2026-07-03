// Phase 6 — Planner Workflow.
// Pure, mock plan generation. Phase 6 is UI-first: the "planner" does not call a
// real model. It poses deterministic clarifying questions, then synthesizes an
// execution plan from the task and the user's answers. No execution, no Git.

import type { ExecutionPlan, PlanningQuestion, Task } from "../../domain/models";

const QUESTION_TEMPLATES: Array<{ id: string; prompt: (task: Task) => string }> = [
  { id: "goal", prompt: (task) => `What is the primary user goal for "${task.title}"?` },
  { id: "scope", prompt: () => "Which modules or files are expected to change? (comma-separated)" },
  { id: "deps", prompt: () => "Are there external dependencies (APIs, database, services)?" },
  { id: "acceptance", prompt: () => "What are the acceptance criteria for this task to be considered done?" },
  { id: "constraints", prompt: () => "Any performance, security, or compatibility constraints to respect?" },
];

export function generateQuestions(task: Task): PlanningQuestion[] {
  return QUESTION_TEMPLATES.map((template) => ({ id: template.id, prompt: template.prompt(task), answer: "" }));
}

function answerOf(questions: PlanningQuestion[], id: string): string {
  return questions.find((question) => question.id === id)?.answer.trim() ?? "";
}

function toFileList(value: string): string[] {
  return value
    .split(/[\n,]+/)
    .map((item) => item.trim())
    .filter(Boolean);
}

// Synthesizes an execution plan from the task and answered questions.
export function generatePlan(task: Task, questions: PlanningQuestion[]): ExecutionPlan {
  const goal = answerOf(questions, "goal");
  const scope = answerOf(questions, "scope");
  const deps = answerOf(questions, "deps");
  const acceptance = answerOf(questions, "acceptance");
  const constraints = answerOf(questions, "constraints");

  const files = toFileList(scope);
  const risks: string[] = [];
  if (constraints) risks.push(constraints);
  if (deps) risks.push(`External dependency: ${deps}`);
  if (risks.length === 0) risks.push("No major risks identified.");

  return {
    id: `plan-${task.id}`,
    taskId: task.id,
    summary: `Plan for ${task.title}.${goal ? ` Goal: ${goal}` : ""}`,
    steps: [
      { id: "1", title: "Analyze requirements and data model", description: goal || task.description || "Clarify scope and expected behavior.", status: "pending" },
      { id: "2", title: "Design and implement changes", description: scope ? `Touch: ${scope}` : "Implement the core changes in the affected modules.", status: "pending" },
      { id: "3", title: "Add tests and verify", description: acceptance || "Cover new behavior with tests and verify acceptance criteria.", status: "pending" },
      { id: "4", title: "Review and update memory", description: "Prepare the diff for review and record key decisions.", status: "pending" },
    ],
    risks,
    filesToTouch: files.length ? files : ["src/"],
    testStrategy: [
      acceptance ? `Verify: ${acceptance}` : "Unit tests for new logic",
      "Manual smoke test of the affected flow",
    ],
    approvalStatus: "pending",
  };
}
