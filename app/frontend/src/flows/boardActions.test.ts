import { describe, expect, it } from "vitest";
import { taskMoveActionForStatus } from "./boardActions";

describe("board actions", () => {
  it("starts an implementation turn when a task moves to In Progress", () => {
    expect(taskMoveActionForStatus("in_progress")).toBe("start-turn");
  });

  it("uses a status update for non-running board moves", () => {
    expect(taskMoveActionForStatus("waiting")).toBe("update-status");
    expect(taskMoveActionForStatus("done")).toBe("update-status");
  });
});
