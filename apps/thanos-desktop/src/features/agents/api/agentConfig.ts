// Phase 4 — Agent Configuration.
// Local persistence for workflow-step configuration and per-provider enable
// overrides. Configuration over hardcoding: user choices survive reloads.

import type { AgentProvider, WorkflowStepConfig } from "../../../domain/models";

const STEPS_KEY = "thanos.agentConfig.steps.v1";
const PROVIDERS_KEY = "thanos.agentConfig.providerOverrides.v1";

export type ProviderOverride = { enabled: boolean };

// Loads persisted steps, overlaid onto the provided defaults so newly added
// fields (e.g. args) always exist even for older saved configs.
export function loadWorkflowSteps(defaults: WorkflowStepConfig[]): WorkflowStepConfig[] {
  try {
    const raw = localStorage.getItem(STEPS_KEY);
    if (!raw) return defaults;
    const saved = JSON.parse(raw) as WorkflowStepConfig[];
    return defaults.map((base) => {
      const override = saved.find((item) => item.id === base.id);
      if (!override) return base;
      return {
        ...base,
        ...override,
        args: override.args ?? base.args ?? [],
        env: override.env ?? base.env ?? {},
        permissions: override.permissions ?? base.permissions ?? [],
      };
    });
  } catch {
    return defaults;
  }
}

export function saveWorkflowSteps(steps: WorkflowStepConfig[]) {
  localStorage.setItem(STEPS_KEY, JSON.stringify(steps));
}

export function loadProviderOverrides(): Record<string, ProviderOverride> {
  try {
    const raw = localStorage.getItem(PROVIDERS_KEY);
    return raw ? (JSON.parse(raw) as Record<string, ProviderOverride>) : {};
  } catch {
    return {};
  }
}

export function saveProviderOverride(id: string, override: ProviderOverride) {
  const all = loadProviderOverrides();
  all[id] = override;
  localStorage.setItem(PROVIDERS_KEY, JSON.stringify(all));
}

// Pure merge: apply saved enable overrides onto freshly detected providers.
export function applyProviderOverrides(
  detected: AgentProvider[],
  overrides: Record<string, ProviderOverride>,
): AgentProvider[] {
  return detected.map((provider) =>
    overrides[provider.id] ? { ...provider, enabled: overrides[provider.id].enabled } : provider,
  );
}
