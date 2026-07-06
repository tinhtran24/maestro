import { ArrowRight, Play, Square } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import type { NativeTerminalSession, Workspace } from "../app/types";
import { onWailsEvent, startNativeTerminal, stopNativeTerminal } from "../services/wails";
import { Panel } from "../shared/Panel";

export function AgentGraphView({ workspace }: { workspace: Workspace }) {
  const [session, setSession] = useState<NativeTerminalSession | null>(null);
  const [output, setOutput] = useState<string[]>([]);
  const installed = useMemo(() => workspace.providers.filter((provider) => provider.status === "installed" && provider.supportsRun), [workspace.providers]);

  useEffect(() => {
    const offOutput = onWailsEvent<{ id: string; stream: string; data: string }>("terminal:output", (event) => {
      setOutput((current) => [...current, `[${event.stream}] ${event.data}`]);
    });
    const offExit = onWailsEvent<{ id: string; status: string; code: number }>("terminal:exit", (event) => {
      setOutput((current) => [...current, `[exit] ${event.status} (${event.code})\n`]);
      setSession((current) => current && current.id === event.id ? { ...current, status: event.status } : current);
    });
    return () => {
      offOutput();
      offExit();
    };
  }, []);

  async function runProvider(providerId: string, command: string, name: string) {
    setOutput([]);
    const next = await startNativeTerminal({
      providerId,
      command,
      args: ["--version"],
      cwd: workspace.path,
      label: `${name} --version`,
    });
    if (next) setSession(next);
  }

  async function stopSession() {
    if (!session) return;
    await stopNativeTerminal(session.id);
  }

  return (
    <div className="view-grid">
      <div className="split-view">
      <Panel title="Agents" meta="harness routed">
        <div className="agent-list">
          {workspace.agents.length === 0 ? <p className="empty">No real agent roles loaded. Select a workspace to inspect providers.</p> : null}
          {workspace.agents.map((agent) => (
            <article key={agent.id} className="agent-card">
              <strong>{agent.role}</strong>
              <span>{agent.harness} · {agent.model}</span>
              <p>{agent.capabilities.join(", ")}</p>
            </article>
          ))}
        </div>
      </Panel>
      <Panel title="Flows" meta="editable graph">
        <div className="flow-list">
          {workspace.flows.length === 0 ? <p className="empty">No flows loaded from the real provider.</p> : null}
          {workspace.flows.map((flow) => (
            <article key={flow.id} className="flow-card">
              <h3>{flow.name}</h3>
              <div className="flow-steps">
                {flow.steps.map((step, index) => (
                  <span key={`${flow.id}-${step}`}>
                    {step}
                    {index < flow.steps.length - 1 ? <ArrowRight size={14} /> : null}
                  </span>
                ))}
              </div>
            </article>
          ))}
        </div>
      </Panel>
      </div>
      <Panel title="Native Terminal" meta="host process">
        <div className="terminal-runner">
          <div className="provider-actions">
            {installed.length === 0 ? <p className="empty">No installed runnable providers detected on PATH.</p> : null}
            {installed.map((provider) => (
              <button key={provider.id} onClick={() => runProvider(provider.id, provider.command, provider.name)}>
                <Play size={15} /> {provider.name}
              </button>
            ))}
            <button disabled={!session || session.status !== "running"} onClick={stopSession}>
              <Square size={15} /> Stop
            </button>
          </div>
          <pre className="terminal-output">{output.length ? output.join("") : "Run an installed provider to stream native terminal output here. No auth token is injected by Thanos."}</pre>
        </div>
      </Panel>
    </div>
  );
}
