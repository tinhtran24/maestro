import { describe, expect, it } from "vitest";
import { classifyUrl, extractTask, toCreateInput, type CapturedAttachment } from "./quickCapture";

describe("quickCapture — extractTask", () => {
  it("maps a cart prompt to a shopping cart task", () => {
    const task = extractTask("Build a shopping cart with guest checkout");
    expect(task.title.value).toBe("Implement Shopping Cart");
    expect(task.priority.value).toBe("P1");
    expect(task.labels.value).toContain("cart");
    expect(task.acceptanceCriteria.length).toBeGreaterThan(0);
  });

  it("maps a login prompt to a P0 auth task", () => {
    const task = extractTask("Fix the login bug, OAuth redirects incorrectly");
    expect(task.title.value).toBe("Fix Authentication Flow");
    expect(task.priority.value).toBe("P0");
    expect(task.labels.value).toContain("auth");
  });

  it("keeps confidence within [0,100] and rises with richer input", () => {
    const sparse = extractTask("cart");
    const rich = extractTask("Build a shopping cart with add, remove, quantity update and checkout");
    for (const task of [sparse, rich]) {
      expect(task.title.confidence).toBeGreaterThanOrEqual(0);
      expect(task.title.confidence).toBeLessThanOrEqual(100);
    }
    expect(rich.title.confidence).toBeGreaterThan(sparse.title.confidence);
  });

  it("is deterministic — identical input yields identical output", () => {
    expect(extractTask("Add dark mode")).toEqual(extractTask("Add dark mode"));
  });

  it("falls back to a generic task with open questions on empty input", () => {
    const task = extractTask("");
    expect(task.title.value).toBe("New Task");
    expect(task.feature.value).toBe("General");
    expect(task.openQuestions.length).toBeGreaterThan(0);
  });

  it("adds attachment kinds to labels and boosts confidence", () => {
    const attachments: CapturedAttachment[] = [
      { id: "a1", kind: "figma", label: "Design", url: "https://figma.com/x" },
    ];
    const withAttachment = extractTask("Add search feature", attachments);
    const without = extractTask("Add search feature");
    expect(withAttachment.labels.value).toContain("figma");
    expect(withAttachment.title.confidence).toBeGreaterThan(without.title.confidence);
  });
});

describe("quickCapture — classifyUrl", () => {
  it("classifies known providers and defaults to link", () => {
    expect(classifyUrl("https://www.figma.com/file/abc")).toBe("figma");
    expect(classifyUrl("https://github.com/o/r/issues/1")).toBe("github");
    expect(classifyUrl("https://acme.atlassian.net/browse/T-1")).toBe("jira");
    expect(classifyUrl("https://example.com")).toBe("link");
  });
});

describe("quickCapture — toCreateInput", () => {
  it("maps labels to tags and folds acceptance criteria + notes into the description", () => {
    const input = toCreateInput(extractTask("Build a shopping cart"));
    expect(input.priority).toBe("P1");
    expect(input.tags).toContain("cart");
    expect(input.description).toContain("Acceptance Criteria");
    expect(input.description).toContain("Technical Notes");
    expect(input.title).toBe("Implement Shopping Cart");
  });

  it("falls back to tags [new] when there are no labels", () => {
    const extracted = extractTask("something vague");
    extracted.labels.value = [];
    expect(toCreateInput(extracted).tags).toEqual(["new"]);
  });
});
