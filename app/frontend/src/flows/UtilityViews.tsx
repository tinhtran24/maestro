import { CalendarClock, MessageSquare, PenTool, Settings } from "lucide-react";
import { useState } from "react";
import type { Workspace } from "../app/types";
import { saveAutomation, upsertRoutine } from "../services/wails";
import { Panel } from "../shared/Panel";

export function ChatView() {
  return (
    <Panel title="Chat" meta="workspace context">
      <div className="chat-page">
        <p>Conversation is the entry point for rough ideas. Promote useful turns into specs before dispatch.</p>
        <textarea placeholder="Ask Thanos to explore an idea..." />
      </div>
    </Panel>
  );
}

export function RoutinesView({ workspace, onReload }: { workspace: Workspace; onReload: () => Promise<void> }) {
  const [name, setName] = useState("");
  const [prompt, setPrompt] = useState("");
  const [schedule, setSchedule] = useState("manual");

  async function submitRoutine(event: React.FormEvent) {
    event.preventDefault();
    if (!workspace.path || !name.trim()) return;
    await upsertRoutine({ root: workspace.path, name, prompt, schedule, flow: "implement", enabled: true });
    setName("");
    setPrompt("");
    await onReload();
  }

  return (
    <Panel title="Routines" meta="scheduled prompts">
      <form className="routine-form" onSubmit={submitRoutine}>
        <input value={name} onChange={(event) => setName(event.target.value)} placeholder="Routine name" />
        <input value={schedule} onChange={(event) => setSchedule(event.target.value)} placeholder="manual, daily, weekly..." />
        <textarea value={prompt} onChange={(event) => setPrompt(event.target.value)} placeholder="Prompt to spawn tasks from..." />
        <button disabled={!workspace.path}><CalendarClock size={16} /> Save Routine</button>
      </form>
      <div className="routine-list">
        {workspace.routines.length === 0 ? <p className="empty">No routines saved.</p> : null}
        {workspace.routines.map((routine) => (
          <article key={routine.id}>
            <strong>{routine.name}</strong>
            <span>{routine.schedule} · {routine.enabled ? "enabled" : "disabled"} · {routine.flow}</span>
            <p>{routine.prompt}</p>
          </article>
        ))}
      </div>
    </Panel>
  );
}

export function WhiteboardView() {
  return (
    <Panel title="Whiteboard" meta="freeform">
      <div className="placeholder"><PenTool size={30} /> Sketch architecture, flows, and notes without dispatching agents.</div>
    </Panel>
  );
}

export function AnalyticsView({ workspace }: { workspace: Workspace }) {
  const total = workspace.tasks.reduce((sum, task) => sum + task.usageUsd, 0);
  return (
    <div className="stats">
      <Panel title="Usage" meta="mock data"><strong className="big">${total.toFixed(2)}</strong><p>Total task spend tracked in this workspace.</p></Panel>
      <Panel title="Throughput" meta="mock data"><strong className="big">{workspace.tasks.length}</strong><p>Visible tasks across active board states.</p></Panel>
      <Panel title="Review Pauses" meta="mock data"><strong className="big">1</strong><p>Tasks waiting for human oversight.</p></Panel>
    </div>
  );
}

export function SettingsView({ runtime, workspace, onReload }: { runtime: string; workspace: Workspace; onReload: () => Promise<void> }) {
  const installed = workspace.providers.filter((provider) => provider.status === "installed");
  async function toggle(key: keyof Workspace["automation"]) {
    if (!workspace.path) return;
    await saveAutomation({ root: workspace.path, automation: { ...workspace.automation, [key]: !workspace.automation[key] } });
    await onReload();
  }
  return (
    <Panel title="Settings" meta="local machine">
      <div className="settings-list">
        <p><Settings size={16} /> Runtime: {runtime}</p>
        <p><MessageSquare size={16} /> Installed providers: {installed.length ? installed.map((provider) => provider.name).join(", ") : "None detected"}</p>
        <div className="automation-toggles">
          {(["autoImplement", "autoTest", "autoSubmit", "autoRetry"] as const).map((key) => (
            <button key={key} onClick={() => toggle(key)} className={workspace.automation[key] ? "enabled" : ""}>
              {key}: {workspace.automation[key] ? "on" : "off"}
            </button>
          ))}
        </div>
        <div className="provider-table">
          {workspace.providers.map((provider) => (
            <article key={provider.id}>
              <strong>{provider.name}</strong>
              <span>{provider.status}</span>
              <code>{provider.path || provider.command}</code>
              <small>{provider.version || provider.setupHint}</small>
            </article>
          ))}
        </div>
        {workspace.diagnostics.map((item) => (
          <p key={`${item.kind}-${item.message}`} className="diagnostic">{item.kind}: {item.message}</p>
        ))}
      </div>
    </Panel>
  );
}

export function DocsView() {
  return (
    <Panel title="Docs" meta="local">
      <article className="document">
        <h2>Architecture direction</h2>
        <p>Thanos is a Wails desktop app with a host-native Go backend and a React workbench frontend.</p>
        <p>Wallfacer-inspired concepts are represented as surfaces: board, plan, mission control, agent graph, routines, analytics, and oversight.</p>
      </article>
    </Panel>
  );
}
