import { describe, expect, it } from "vitest";
import type { ExecutionPlan, MemoryNode, Task } from "../../domain/models";
import { memoryFromReview, relatedMemory, searchMemory } from "./memorySearch";

function node(id: string, type: MemoryNode["type"], title: string, content: string, links: string[] = []): MemoryNode {
  return { id, projectId: "P-1", type, title, content, links, createdAt: "2024-05-18T00:00:00Z" };
}

const nodes: MemoryNode[] = [
  node("mem-1", "decision", "Cart System Design", "Frontend-first cart with backend sync.", ["mem-3"]),
  node("mem-2", "architecture", "Database Schema Guidelines", "Indexed foreign keys and rollback migrations."),
  node("mem-3", "feature", "Shopping Cart Requirements", "Add, remove, update, persist."),
];

describe("memory — searchMemory", () => {
  it("returns all nodes for an empty query", () => {
    expect(searchMemory(nodes, "   ")).toHaveLength(3);
  });

  it("matches nodes that contain every query token", () => {
    const results = searchMemory(nodes, "cart");
    expect(results.map((n) => n.id).sort()).toEqual(["mem-1", "mem-3"]);
  });

  it("matches on type and returns nothing when a token is absent", () => {
    expect(searchMemory(nodes, "architecture").map((n) => n.id)).toEqual(["mem-2"]);
    expect(searchMemory(nodes, "cart kubernetes")).toHaveLength(0);
  });
});

describe("memory — relatedMemory", () => {
  it("prefers explicitly linked nodes, then same-type nodes, excluding self", () => {
    const related = relatedMemory(nodes, nodes[0]);
    expect(related.map((n) => n.id)).toContain("mem-3"); // linked
    expect(related.every((n) => n.id !== "mem-1")).toBe(true);
  });

  it("surfaces a back-linked node even if the current node does not link it", () => {
    const extra = [...nodes, node("mem-4", "bug", "Cart sync bug", "Race on checkout.", ["mem-1"])];
    expect(relatedMemory(extra, extra[0]).map((n) => n.id)).toContain("mem-4");
  });
});

describe("memory — memoryFromReview", () => {
  const task = { id: "T-106", featureId: "P-1", title: "Shopping Cart" } as Task;
  const plan = { summary: "Build the cart.", filesToTouch: ["src/cart.ts"] } as ExecutionPlan;

  it("synthesizes a decision node linked to token-overlapping memory", () => {
    const mem = memoryFromReview(task, plan, nodes, "2026-07-04T00:00:00Z");
    expect(mem.id).toBe("mem-t-106");
    expect(mem.type).toBe("decision");
    expect(mem.title).toContain("Shopping Cart");
    expect(mem.content).toContain("Build the cart.");
    // "Cart" overlaps mem-1 and mem-3 titles.
    expect(mem.links.sort()).toEqual(["mem-1", "mem-3"]);
  });
});
