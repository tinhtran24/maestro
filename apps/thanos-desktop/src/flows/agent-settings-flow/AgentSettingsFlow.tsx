import { InstalledAgentsPanel } from "./InstalledAgentsPanel";
import { WorkflowStepSettingsPanel } from "./WorkflowStepSettingsPanel";
import { AgentCommandTestPanel } from "./AgentCommandTestPanel";
import { useAgentSettingsFlow } from "./useAgentSettingsFlow";

export function AgentSettingsFlow({ tab = "agents" }: { tab?: "agents" | "workflow" }) {
  const flow = useAgentSettingsFlow();
  return (
    <section className="grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] bg-bg-app p-4">
      <header className="mb-4">
        <h1 className="text-lg font-semibold">{tab === "workflow" ? "Workflow Steps" : "Installed Agents"}</h1>
        <p className="text-sm text-text-muted">Use local installed tools. Thanos never installs agents automatically.</p>
      </header>
      <div className="grid min-h-0 grid-cols-[minmax(0,1fr)_22rem] gap-4">
        {tab === "workflow" ? <WorkflowStepSettingsPanel flow={flow} /> : <InstalledAgentsPanel flow={flow} />}
        <AgentCommandTestPanel flow={flow} />
      </div>
    </section>
  );
}
