// Phase 4 — Agent Configuration.
// Mock installed-agent detection. Phase 4 is detection-only mock: this never
// runs a real PATH lookup or executes anything. It returns the supported agent
// catalog with deterministic mock statuses so the settings UI is fully usable
// without a backend. Real `which`-based detection is a later phase.

import type { AgentProvider } from "../../../domain/models";

type CatalogEntry = {
  id: string;
  name: string;
  command: string;
  type: AgentProvider["type"];
  status: AgentProvider["status"];
  version?: string;
  path?: string;
  setupHint: string;
};

// The supported providers (AGENTS.md). This is the catalog of *available*
// agents — not a hardcoded per-step assignment. Users choose which agent runs
// each workflow step in the settings UI.
export const AGENT_CATALOG: CatalogEntry[] = [
  { id: "claude-code", name: "Claude Code", command: "claude", type: "cli", status: "installed", version: "1.0.0", path: "/usr/local/bin/claude", setupHint: "" },
  { id: "codex", name: "Codex", command: "codex", type: "cli", status: "installed", version: "0.9.2", path: "/usr/local/bin/codex", setupHint: "" },
  { id: "gemini-cli", name: "Gemini CLI", command: "gemini", type: "cli", status: "not_found", setupHint: "Install the Gemini CLI and ensure `gemini` is on your PATH." },
  { id: "opencode", name: "OpenCode", command: "opencode", type: "cli", status: "needs_setup", version: "0.3.1", path: "/opt/opencode/bin/opencode", setupHint: "Run `opencode auth login` to finish setup." },
  { id: "cursor-agent", name: "Cursor Agent", command: "cursor-agent", type: "acp", status: "not_found", setupHint: "Install Cursor and enable the agent CLI." },
  { id: "aider", name: "Aider", command: "aider", type: "cli", status: "installed", version: "0.60.0", path: "/usr/local/bin/aider", setupHint: "" },
  { id: "goose", name: "Goose", command: "goose", type: "shell", status: "not_found", setupHint: "Install Goose (`goose`) to enable this agent." },
  { id: "custom", name: "Custom Command", command: "", type: "shell", status: "needs_setup", setupHint: "Set a custom command to run as an agent." },
];

export function mockDetectAgents(): AgentProvider[] {
  return AGENT_CATALOG.map((entry) => ({
    id: entry.id,
    name: entry.name,
    command: entry.command,
    detectedPath: entry.path,
    status: entry.status,
    version: entry.version,
    type: entry.type,
    enabled: entry.status === "installed",
    setupHint: entry.setupHint,
  }));
}
