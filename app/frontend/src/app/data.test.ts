import { describe, expect, it } from "vitest";
import { navItems } from "../shared/Layout";

describe("fresh Wails workbench shell", () => {
  it("exposes the Wallfacer-inspired primary surfaces", () => {
    expect(navItems.map((item) => item.id)).toEqual([
      "board",
      "plan",
      "mission",
      "agent-graph",
      "chat",
      "routines",
      "whiteboard",
      "analytics",
      "settings",
    ]);
  });
});
