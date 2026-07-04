import { describe, expect, it } from "vitest";
import type { GitDiff, Task, TestRun } from "../../domain/models";
import { buildReviewChecklist, canFinish, checklistComplete, mockTestRun } from "./reviewChecklist";

function makeTask(overrides: Partial<Task> = {}): Task {
  return {
    id: "T-1",
    featureId: "F-1",
    title: "Shopping Cart",
    description: "Add cart",
    status: "in_review",
    priority: "P1",
    assignedAgent: "Reviewer",
    executorProfile: "claude-local",
    worktreePath: ".thanos/worktrees/t-1",
    branchName: "thanos/t-1",
    planApproved: true,
    reviewApproved: false,
    testsPassed: true,
    updatedAt: "",
    tags: [],
    progress: 0,
    ...overrides,
  };
}

const diff: GitDiff = {
  taskId: "T-1",
  summary: "2 files changed",
  changedFiles: [{ path: "src/cart.ts", status: "modified" }, { path: "src/api.ts", status: "modified" }],
  patch: "",
};

describe("reviewer — buildReviewChecklist", () => {
  it("marks diff + tests + plan done when artifacts are present", () => {
    const items = buildReviewChecklist(makeTask(), diff, mockTestRun(makeTask(), "npm test"));
    expect(items.find((i) => i.id === "diff")?.done).toBe(true);
    expect(items.find((i) => i.id === "tests")?.done).toBe(true);
    expect(items.find((i) => i.id === "plan")?.done).toBe(true);
    expect(checklistComplete(items)).toBe(true);
  });

  it("is incomplete when the diff is missing or tests have not passed", () => {
    expect(checklistComplete(buildReviewChecklist(makeTask({ testsPassed: false })))).toBe(false);
    expect(buildReviewChecklist(makeTask()).find((i) => i.id === "diff")?.done).toBe(false);
  });

  it("marks tests not done when the test run failed", () => {
    const failed: TestRun = { taskId: "T-1", command: "npm test", status: "failed", stdout: "", stderr: "boom", code: 1 };
    expect(buildReviewChecklist(makeTask(), diff, failed).find((i) => i.id === "tests")?.done).toBe(false);
  });
});

describe("reviewer — canFinish", () => {
  it("requires both review approval and passing tests", () => {
    expect(canFinish(makeTask({ reviewApproved: true, testsPassed: true }))).toBe(true);
    expect(canFinish(makeTask({ reviewApproved: true, testsPassed: false }))).toBe(false);
    expect(canFinish(makeTask({ reviewApproved: false, testsPassed: true }))).toBe(false);
  });
});

describe("reviewer — mockTestRun", () => {
  it("produces a passing run that echoes the command", () => {
    const run = mockTestRun(makeTask(), "npm test");
    expect(run.status).toBe("passed");
    expect(run.stdout).toContain("npm test");
    expect(run.code).toBe(0);
  });
});
