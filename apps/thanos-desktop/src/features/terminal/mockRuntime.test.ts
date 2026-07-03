import { describe, expect, it } from "vitest";
import { scriptFor, sessionIdFor, terminalLabel, type TerminalStep } from "./mockRuntime";

const STEPS: TerminalStep[] = ["planning", "coding", "review", "testing"];

describe("mock terminal runtime — scriptFor", () => {
  it("produces non-empty output for every step", () => {
    for (const step of STEPS) {
      const lines = scriptFor(step, "claude", "Shopping Cart");
      expect(lines.length).toBeGreaterThan(0);
      expect(lines.every((line) => typeof line === "string")).toBe(true);
    }
  });

  it("echoes the command for a shell/test session", () => {
    const lines = scriptFor("testing", "npm test", "Any");
    expect(lines[0]).toContain("npm test");
  });

  it("includes the task title in the planning session", () => {
    const lines = scriptFor("planning", "claude", "Shopping Cart");
    expect(lines.some((line) => line.includes("Shopping Cart"))).toBe(true);
  });
});

describe("mock terminal runtime — identifiers + labels", () => {
  it("derives a deterministic session id per task + step", () => {
    expect(sessionIdFor("T-1", "planning")).toBe("sess-T-1-planning");
    expect(sessionIdFor("T-1", "planning")).toBe(sessionIdFor("T-1", "planning"));
  });

  it("labels a session with its step and provider", () => {
    expect(terminalLabel("planning", "Claude Code")).toBe("Planning · Claude Code");
    expect(terminalLabel("testing", "Shell")).toBe("Tests · Shell");
    expect(terminalLabel(undefined, "codex")).toBe("session · codex");
  });
});
