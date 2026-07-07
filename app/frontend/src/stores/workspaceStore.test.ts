import { describe, expect, it } from "vitest";
import { emptyWorkspace, preserveWorkspaceIdentity } from "./workspaceStore";

describe("workspace store helpers", () => {
  it("keeps registry identity while refreshing workspace data", () => {
    const previous = {
      ...emptyWorkspace,
      workspaceId: "ws-1",
      dataKey: "data-1",
      folders: [
        { id: "repo", path: "/repo", label: "repo" },
        { id: "docs", path: "/docs", label: "docs" },
      ],
    };
    const loaded = {
      ...emptyWorkspace,
      workspaceId: "",
      dataKey: "",
      path: "/repo",
      folders: [{ id: "repo", path: "/repo", label: "repo" }],
      name: "repo",
    };

    expect(preserveWorkspaceIdentity(previous, loaded)).toMatchObject({
      workspaceId: "ws-1",
      dataKey: "data-1",
      folders: [
        { id: "repo", path: "/repo", label: "repo" },
        { id: "docs", path: "/docs", label: "docs" },
      ],
      path: "/repo",
      name: "repo",
    });
  });
});
