import { beforeEach, describe, expect, it } from "vitest";
import type { Task } from "../../domain/models";
import { agentTypeFor, hasBackend, startAgentStep } from "./agentRuntime";
import { sessionIdFor } from "./mockRuntime";
import { useWorkbenchStore } from "../../state/workbenchStore";

function makeTask(overrides: Partial<Task> = {}): Task {
  return {
    id: "T-1",
    featureId: "F-1",
    title: "Shopping Cart",
    description: "",
    status: "ready",
    priority: "P1",
    assignedAgent: "Codex",
    executorProfile: "codex-local",
    worktreePath: ".thanos/worktrees/t-1",
    branchName: "thanos/t-1",
    planApproved: true,
    reviewApproved: false,
    testsPassed: false,
    updatedAt: "",
    tags: [],
    progress: 0,
    ...overrides,
  };
}

describe("agentRuntime — agentTypeFor", () => {
  it("maps workflow steps to agent roles", () => {
    expect(agentTypeFor("planning")).toBe("planner");
    expect(agentTypeFor("coding")).toBe("coder");
    expect(agentTypeFor("review")).toBe("reviewer");
    expect(agentTypeFor("testing")).toBe("tester");
  });
});

describe("agentRuntime — startAgentStep (no backend)", () => {
  beforeEach(() => {
    useWorkbenchStore.setState({ sessions: [], activeSessionByTask: {} });
  });

  it("falls back to the labeled simulation when no Tauri backend is present", async () => {
    expect(hasBackend()).toBe(false); // vitest/node has no __TAURI_INTERNALS__
    const id = await startAgentStep(makeTask(), "planning");
    expect(id).toBe(sessionIdFor("T-1", "planning"));
    const session = useWorkbenchStore.getState().sessions.find((item) => item.id === id);
    expect(session?.output.join("")).toContain("[simulated");
  });

  it("blocks coding until the plan is approved and explains why", async () => {
    const id = await startAgentStep(makeTask({ planApproved: false }), "coding");
    const session = useWorkbenchStore.getState().sessions.find((item) => item.id === id);
    expect(session?.status).toBe("failed");
    expect(session?.output.join("")).toContain("plan");
  });
});
