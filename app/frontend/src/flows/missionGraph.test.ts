import { describe, expect, it } from "vitest";
import type { Workspace } from "../app/types";
import { emptyWorkspace } from "../stores/workspaceStore";
import { buildMissionGraph } from "./missionGraph";

const baseTask = {
  schema_version: 1,
  prompt: "",
  flow: "implement",
  agent: "codex",
  branch: "",
  worktree: "",
  updatedAt: "",
  usageUsd: 0,
  archived: false,
  deleted: false,
  tombstone: false,
  blocked: false,
  promptHistory: [],
  feedbackHistory: [],
  retryHistory: [],
  turns: [],
  testsPassed: false,
  failureCategory: "",
  createdAt: "",
};

describe("mission graph", () => {
  it("builds typed edges, blocked set, and critical path from workspace entities", () => {
    const workspace: Workspace = {
      ...emptyWorkspace,
      specs: [
        {
          id: "checkout",
          title: "Checkout",
          state: "validated",
          path: "specs/checkout.md",
          body: "",
          updatedAt: "",
          children: [],
        },
      ],
      tasks: [
        { ...baseTask, id: "T-1", title: "Checkout", status: "done", dependencies: [] },
        { ...baseTask, id: "T-2", title: "Payment", status: "waiting", dependencies: ["T-1"] },
        { ...baseTask, id: "T-3", title: "Receipt", status: "backlog", dependencies: ["T-2"], blocked: true, prompt: "Routine: routine-nightly" },
      ],
      routines: [{ id: "routine-nightly", name: "Nightly", prompt: "", flow: "implement", schedule: "daily", enabled: true, runCount: 1, failureCount: 0, updatedAt: "" }],
      providers: [{ id: "codex", name: "Codex", command: "codex", status: "installed", type: "cli", setupHint: "", supportsRun: true }],
      flows: [{ id: "implement", name: "Implement", steps: ["Implementation"], readOnly: true }],
    };

    const graph = buildMissionGraph(workspace);

    expect(graph.nodes.map((node) => node.id)).toContain("spec:specs/checkout.md");
    expect(graph.edges).toEqual(expect.arrayContaining([
      expect.objectContaining({ from: "spec:specs/checkout.md", to: "task:T-1", kind: "dispatches" }),
      expect.objectContaining({ from: "task:T-3", to: "task:T-2", kind: "blocked_by" }),
      expect.objectContaining({ from: "routine:routine-nightly", to: "task:T-3", kind: "produced" }),
    ]));
    expect(graph.blockedTaskIds).toEqual(["T-3"]);
    expect(graph.criticalPath).toEqual(["T-1", "T-2", "T-3"]);
  });
});
