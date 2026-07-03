import { describe, expect, it } from "vitest";
import type { Task } from "../domain/models";
import { availableTransitions, canTransition, MAIN_FLOW, statusLabel, transitionTask } from "./taskMachine";

function makeTask(overrides: Partial<Task> = {}): Task {
  return {
    id: "T-1",
    featureId: "F-1",
    title: "Test task",
    description: "",
    status: "backlog",
    priority: "P2",
    assignedAgent: "Agent",
    executorProfile: "codex-local",
    worktreePath: "",
    branchName: "",
    planApproved: false,
    reviewApproved: false,
    testsPassed: false,
    updatedAt: "",
    tags: [],
    progress: 0,
    ...overrides,
  };
}

describe("taskMachine — allowed transitions", () => {
  it("permits the canonical happy path step by step", () => {
    expect(canTransition(makeTask({ status: "backlog" }), "planning")).toBeNull();
    expect(canTransition(makeTask({ status: "planning" }), "waiting_approval")).toBeNull();
    expect(canTransition(makeTask({ status: "waiting_approval", planApproved: true }), "ready")).toBeNull();
    expect(
      canTransition(makeTask({ status: "ready", worktreePath: ".w", branchName: "b" }), "running"),
    ).toBeNull();
    expect(canTransition(makeTask({ status: "running" }), "in_review")).toBeNull();
    expect(
      canTransition(makeTask({ status: "in_review", reviewApproved: true, testsPassed: true }), "done"),
    ).toBeNull();
  });
});

describe("taskMachine — invalid transitions are rejected", () => {
  it("rejects transitions not present in the table", () => {
    expect(canTransition(makeTask({ status: "backlog" }), "running")).toMatch(/Invalid transition/);
    expect(canTransition(makeTask({ status: "done" }), "planning")).toMatch(/Invalid transition/);
    expect(canTransition(makeTask({ status: "planning" }), "done")).toMatch(/Invalid transition/);
  });

  it("keeps done terminal (no outgoing transitions)", () => {
    expect(availableTransitions(makeTask({ status: "done" }))).toHaveLength(0);
  });
});

describe("taskMachine — gates", () => {
  it("blocks ready until the plan is approved", () => {
    expect(canTransition(makeTask({ status: "waiting_approval", planApproved: false }), "ready")).toMatch(
      /Plan approval is required/,
    );
  });

  it("blocks running until a worktree and branch exist", () => {
    expect(canTransition(makeTask({ status: "ready" }), "running")).toMatch(/isolated worktree and branch/);
  });

  it("blocks done until review approved and tests pass", () => {
    expect(canTransition(makeTask({ status: "in_review", reviewApproved: false, testsPassed: true }), "done")).toMatch(
      /Review approval is required/,
    );
    expect(canTransition(makeTask({ status: "in_review", reviewApproved: true, testsPassed: false }), "done")).toMatch(
      /Passing tests are required/,
    );
  });
});

describe("taskMachine — transitionTask + availableTransitions", () => {
  it("returns the updated task on a valid transition", () => {
    const result = transitionTask(makeTask({ status: "backlog" }), "planning");
    expect(result.ok).toBe(true);
    if (result.ok) expect(result.task.status).toBe("planning");
  });

  it("returns a reason and does not mutate on an invalid transition", () => {
    const result = transitionTask(makeTask({ status: "backlog" }), "done");
    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.reason).toBeTruthy();
  });

  it("annotates each target with allowed + reason", () => {
    const options = availableTransitions(makeTask({ status: "waiting_approval", planApproved: false }));
    const ready = options.find((option) => option.to === "ready");
    expect(ready?.allowed).toBe(false);
    expect(ready?.reason).toMatch(/Plan approval/);
    const planning = options.find((option) => option.to === "planning");
    expect(planning?.allowed).toBe(true);
  });
});

describe("taskMachine — metadata", () => {
  it("orders the main flow and labels statuses", () => {
    expect(MAIN_FLOW[0]).toBe("backlog");
    expect(MAIN_FLOW[MAIN_FLOW.length - 1]).toBe("done");
    expect(statusLabel("in_review")).toBe("In Review");
    expect(statusLabel("running")).toBe("In Progress");
  });
});
