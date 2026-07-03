import { AlertTriangle, CheckCircle2, Play, RefreshCw, ToggleLeft, ToggleRight, XCircle } from "lucide-react";
import type { AgentProvider } from "../../domain/models";
import type { useAgentSettingsFlow } from "./useAgentSettingsFlow";

const statusMeta: Record<AgentProvider["status"], { label: string; className: string; icon: typeof CheckCircle2 }> = {
  installed: { label: "Installed", className: "border-green-success/30 bg-green-success/10 text-green-success", icon: CheckCircle2 },
  not_found: { label: "Not Found", className: "border-red-danger/30 bg-red-danger/10 text-red-danger", icon: XCircle },
  needs_setup: { label: "Needs Setup", className: "border-yellow-warning/30 bg-yellow-warning/10 text-yellow-warning", icon: AlertTriangle },
};

export function InstalledAgentsPanel({ flow }: { flow: ReturnType<typeof useAgentSettingsFlow> }) {
  return (
    <div className="min-h-0 overflow-y-auto rounded-lg border border-slate-800 bg-bg-card">
      <header className="flex items-center justify-between border-b border-slate-800 p-3">
        <div>
          <h2 className="font-semibold">Installed Agents</h2>
          <p className="text-xs text-text-muted">Detected locally (mock). Thanos never installs agents automatically.</p>
        </div>
        <button onClick={flow.refresh} className="inline-flex items-center gap-2 rounded-lg border border-slate-800 px-2 py-2 text-xs text-text-muted hover:text-text-main">
          <RefreshCw size={15} /> Rescan
        </button>
      </header>
      <div className="grid gap-2 p-3">
        {flow.providers.map((agent) => {
          const meta = statusMeta[agent.status];
          const StatusIcon = meta.icon;
          return (
            <article key={agent.id} className="rounded-lg border border-slate-800 bg-slate-950/50 p-3">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <p className="font-medium">{agent.name}</p>
                    <span className={`inline-flex items-center gap-1 rounded-md border px-2 py-0.5 text-[11px] ${meta.className}`}>
                      <StatusIcon size={11} /> {meta.label}
                    </span>
                    <span className="rounded-md border border-slate-700 px-1.5 py-0.5 text-[10px] uppercase tracking-wide text-text-muted">{agent.type}</span>
                  </div>
                  <p className="mt-1 truncate text-xs text-text-muted">{agent.detectedPath || agent.setupHint || "Path unavailable"}</p>
                  <p className="mt-0.5 text-xs text-text-muted">{agent.version ? `v${agent.version}` : "Version unavailable"}</p>
                </div>
                <button
                  onClick={() => flow.updateProvider(agent.id, { enabled: !agent.enabled })}
                  className="shrink-0 text-text-muted"
                  aria-label={agent.enabled ? "Disable agent" : "Enable agent"}
                  title={agent.enabled ? "Enabled" : "Disabled"}
                >
                  {agent.enabled ? <ToggleRight className="text-green-success" /> : <ToggleLeft />}
                </button>
              </div>
              <div className="mt-3 flex flex-wrap items-center gap-2 text-xs">
                <span className="rounded-md border border-slate-700 px-2 py-1 font-mono">{agent.command || "—"}</span>
                <button onClick={() => flow.testRun(agent)} className="inline-flex items-center gap-1 rounded-md border border-slate-700 px-2 py-1 hover:border-slate-500">
                  <Play size={12} /> Test Run
                </button>
                <select
                  value=""
                  onChange={(event) => {
                    const step = flow.workflowSteps.find((item) => item.id === event.target.value);
                    if (step) flow.setStepProvider(step.id, agent);
                    event.currentTarget.value = "";
                  }}
                  className="ml-auto h-7 rounded-md border border-slate-700 bg-slate-950 px-2 text-xs text-text-muted"
                >
                  <option value="">Use for step…</option>
                  {flow.workflowSteps.map((step) => (
                    <option key={step.id} value={step.id}>{step.label}</option>
                  ))}
                </select>
              </div>
            </article>
          );
        })}
      </div>
    </div>
  );
}
