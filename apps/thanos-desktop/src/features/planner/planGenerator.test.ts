import { describe, expect, it } from "vitest";
import type { PlanningQuestion, Task } from "../../domain/models";
import { generatePlan, generateQuestions } from "./planGenerator";

function makeTask(overrides: Partial<Task> = {}): Task {
  return {
    id: "T-1",
    featureId: "F-1",
    title: "Shopping Cart",
    description: "Add cart",
    status: "planning",
    priority: "P1",
    assignedAgent: "Planner",
    executorProfile: "claude-local",
    worktreePath: "",
    branchName: "",
    reviewApproved: false,
    testsPassed: false,
    updatedAt: "",
    tags: [],
    progress: 0,
    ...overrides,
  };
}

function answered(): PlanningQuestion[] {
  return [
    { id: "goal", prompt: "goal?", answer: "Let users buy items" },
    { id: "scope", prompt: "scope?", answer: "src/cart.ts, src/api.ts" },
    { id: "deps", prompt: "deps?", answer: "Stripe API" },
    { id: "acceptance", prompt: "acceptance?", answer: "Cart persists across reloads" },
    { id: "constraints", prompt: "constraints?", answer: "PCI compliance" },
  ];
}

describe("planner — generateQuestions", () => {
  it("produces clarifying questions with empty answers", () => {
    const questions = generateQuestions(makeTask());
    expect(questions.length).toBeGreaterThanOrEqual(4);
    expect(questions.every((q) => q.answer === "")).toBe(true);
    expect(questions.find((q) => q.id === "goal")?.prompt).toContain("Shopping Cart");
  });
});

describe("planner — generatePlan", () => {
  it("synthesizes a plan from answers", () => {
    const plan = generatePlan(makeTask(), answered());
    expect(plan.taskId).toBe("T-1");
    expect(plan.approvalStatus).toBe("pending");
    expect(plan.steps).toHaveLength(4);
    expect(plan.filesToTouch).toEqual(["src/cart.ts", "src/api.ts"]);
    expect(plan.summary).toContain("Let users buy items");
    expect(plan.risks.join(" ")).toContain("PCI compliance");
    expect(plan.risks.join(" ")).toContain("Stripe API");
  });

  it("falls back to a default file list and risk when answers are sparse", () => {
    const sparse: PlanningQuestion[] = [
      { id: "goal", prompt: "", answer: "" },
      { id: "scope", prompt: "", answer: "" },
      { id: "deps", prompt: "", answer: "" },
      { id: "acceptance", prompt: "", answer: "" },
      { id: "constraints", prompt: "", answer: "" },
    ];
    const plan = generatePlan(makeTask(), sparse);
    expect(plan.filesToTouch).toEqual(["src/"]);
    expect(plan.risks).toContain("No major risks identified.");
  });
});
