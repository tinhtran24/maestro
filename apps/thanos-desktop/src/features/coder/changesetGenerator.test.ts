import { describe, expect, it } from "vitest";
import type { ExecutionPlan, Task } from "../../domain/models";
import { generateChangeset } from "./changesetGenerator";

function makeTask(overrides: Partial<Task> = {}): Task {
  return {
    id: "T-1",
    featureId: "F-1",
    title: "Shopping Cart",
    description: "Add cart",
    status: "running",
    priority: "P1",
    assignedAgent: "Codex",
    executorProfile: "codex-local",
    worktreePath: ".thanos/worktrees/t-1",
    branchName: "thanos/t-1-shopping-cart",
    planApproved: true,
    reviewApproved: false,
    testsPassed: false,
    updatedAt: "",
    tags: [],
    progress: 0,
    ...overrides,
  };
}

function makePlan(filesToTouch: string[]): ExecutionPlan {
  return {
    id: "plan-T-1",
    taskId: "T-1",
    summary: "Plan",
    steps: [],
    risks: [],
    filesToTouch,
    testStrategy: [],
    approvalStatus: "approved",
  };
}

describe("coder — generateChangeset", () => {
  it("produces a changed file per planned file with an add/modify status", () => {
    const diff = generateChangeset(makeTask(), makePlan(["src/cart.ts", "src/api.ts"]));
    expect(diff.taskId).toBe("T-1");
    expect(diff.changedFiles.map((file) => file.path)).toEqual(["src/cart.ts", "src/api.ts"]);
    expect(diff.changedFiles.every((file) => file.status === "modified")).toBe(true);
  });

  it("treats extensionless paths and directories as new files", () => {
    const diff = generateChangeset(makeTask(), makePlan(["src/", "src/newModule"]));
    expect(diff.changedFiles.every((file) => file.status === "added")).toBe(true);
    // Added files never remove lines.
    expect(diff.summary).toContain("0 deletions(-)");
  });

  it("is deterministic — identical inputs yield identical output", () => {
    const plan = makePlan(["src/cart.ts", "src/api.ts"]);
    expect(generateChangeset(makeTask(), plan)).toEqual(generateChangeset(makeTask(), plan));
  });

  it("summarizes insertions and deletions across all files", () => {
    const diff = generateChangeset(makeTask(), makePlan(["src/cart.ts"]));
    expect(diff.summary).toMatch(/1 file changed, \d+ insertions\(\+\), \d+ deletions\(-\)/);
    expect(diff.patch).toContain("src/cart.ts");
  });

  it("falls back to a default file when the plan touches nothing", () => {
    const diff = generateChangeset(makeTask(), makePlan([]));
    expect(diff.changedFiles).toHaveLength(1);
    expect(diff.changedFiles[0].path).toBe("src/");
  });
});
