import { describe, expect, it } from "vitest";
import { applyProviderOverrides } from "./agentConfig";
import { AGENT_CATALOG, mockDetectAgents } from "./agentCatalog";

describe("agent catalog — mock detection", () => {
  it("returns every supported provider", () => {
    const providers = mockDetectAgents();
    expect(providers).toHaveLength(AGENT_CATALOG.length);
    expect(providers.map((p) => p.id)).toContain("claude-code");
    expect(providers.map((p) => p.id)).toContain("custom");
  });

  it("assigns only valid statuses and enables installed ones by default", () => {
    for (const provider of mockDetectAgents()) {
      expect(["installed", "not_found", "needs_setup"]).toContain(provider.status);
      expect(provider.enabled).toBe(provider.status === "installed");
    }
  });

  it("exposes a setup hint for providers that are not installed", () => {
    for (const provider of mockDetectAgents()) {
      if (provider.status !== "installed") expect((provider.setupHint ?? "").length).toBeGreaterThan(0);
    }
  });
});

describe("agent config — applyProviderOverrides", () => {
  it("applies saved enable overrides onto detected providers", () => {
    const detected = mockDetectAgents();
    const claude = detected.find((p) => p.id === "claude-code")!;
    expect(claude.enabled).toBe(true);

    const merged = applyProviderOverrides(detected, { "claude-code": { enabled: false } });
    expect(merged.find((p) => p.id === "claude-code")!.enabled).toBe(false);
    // Providers without an override are untouched.
    expect(merged.find((p) => p.id === "codex")!.enabled).toBe(detected.find((p) => p.id === "codex")!.enabled);
  });

  it("does not mutate the input array", () => {
    const detected = mockDetectAgents();
    const before = detected.find((p) => p.id === "claude-code")!.enabled;
    applyProviderOverrides(detected, { "claude-code": { enabled: !before } });
    expect(detected.find((p) => p.id === "claude-code")!.enabled).toBe(before);
  });
});
