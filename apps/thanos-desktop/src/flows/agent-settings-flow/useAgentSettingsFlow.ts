import { useEffect, useState } from "react";
import type { AgentProvider, WorkflowStepId } from "../../domain/models";
import { mockDetectAgents } from "../../features/agents/api/agentCatalog";
import { NativeBackend } from "../../services/nativeBackend";
import { useWorkbenchStore } from "../../state/workbenchStore";

const backend = new NativeBackend();

export type TestResult = { name: string; command: string; ok: boolean; output: string };

export function useAgentSettingsFlow() {
  const providers = useWorkbenchStore((state) => state.agentProviders);
  const setProviders = useWorkbenchStore((state) => state.setAgentProviders);
  const updateProvider = useWorkbenchStore((state) => state.updateAgentProvider);
  const workflowSteps = useWorkbenchStore((state) => state.workflowSteps);
  const updateWorkflowStep = useWorkbenchStore((state) => state.updateWorkflowStep);
  const [testing, setTesting] = useState("");
  const [testResult, setTestResult] = useState<TestResult | null>(null);

  // Phase 4: mock detection. Prefer the native backend when present, otherwise
  // fall back to the mock catalog so the settings UI works without Tauri.
  async function refresh() {
    const detected = await backend.detectAgents();
    setProviders(detected.length ? detected : mockDetectAgents());
  }

  useEffect(() => {
    if (!providers.length) void refresh();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Mock test run — inspects the (mock) detected status only. Never executes.
  async function testRun(provider: AgentProvider) {
    const label = provider.command || provider.name;
    setTesting(label);
    await new Promise((resolve) => window.setTimeout(resolve, 650));
    const ok = provider.status === "installed";
    setTestResult({
      name: provider.name,
      command: label,
      ok,
      output: ok
        ? `${provider.command} --version → ${provider.version ?? "detected"}`
        : `${provider.name}: ${provider.setupHint || "not found on PATH"}`,
    });
    setTesting("");
  }

  // Assign a provider as the agent that runs a given workflow step.
  function setStepProvider(stepId: WorkflowStepId, provider: AgentProvider) {
    updateWorkflowStep(stepId, { provider: provider.name, command: provider.command });
  }

  return { providers, workflowSteps, refresh, updateProvider, updateWorkflowStep, setStepProvider, testRun, testing, testResult };
}
