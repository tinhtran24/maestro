import { ArrowRight, Bot, Maximize2, Play, Send, Square } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import type { AgentMode, AgentSession, Workspace } from "../app/types";
import { listAgentProviders, listAgentSessions, onWailsEvent, resizeNativeTerminal, sendAgentInput, startAgentSession, stopAgentSession } from "../services/wails";
import { Panel } from "../shared/Panel";

const agentModes: AgentMode[] = ["planner", "coding", "review", "debug", "research"];

export function AgentGraphView({ workspace }: { workspace: Workspace }) {
  const [session, setSession] = useState<AgentSession | null>(null);
  const [sessions, setSessions] = useState<AgentSession[]>([]);
  const [providers, setProviders] = useState(workspace.providers);
  const [providerId, setProviderId] = useState(workspace.providers[0]?.id ?? "codex");
  const [mode, setMode] = useState<AgentMode>("coding");
  const [taskId, setTaskId] = useState("");
  const [prompt, setPrompt] = useState("");
  const [customCommand, setCustomCommand] = useState("");
  const [output, setOutput] = useState<string[]>([]);
  const [input, setInput] = useState("");
  const selectedProvider = useMemo(() => providers.find((provider) => provider.id === providerId), [providers, providerId]);
  const selectedTask = useMemo(() => workspace.tasks.find((task) => task.id === taskId), [workspace.tasks, taskId]);
  const canStart = Boolean(workspace.path && selectedProvider && prompt.trim() && (selectedProvider.status === "installed" || (selectedProvider.id === "custom-local" && customCommand.trim())));

  useEffect(() => {
    const offOutput = onWailsEvent<{ id: string; stream: string; data: string }>("terminal:output", (event) => {
      setOutput((current) => session && event.id === session.terminalId ? [...current, `[${event.stream}] ${event.data}`] : current);
    });
    const offExit = onWailsEvent<{ id: string; status: string; code: number }>("terminal:exit", (event) => {
      setOutput((current) => session && event.id === session.terminalId ? [...current, `[exit] ${event.status} (${event.code})\n`] : current);
      setSession((current) => current && current.terminalId === event.id ? { ...current, status: event.status } : current);
    });
    return () => {
      offOutput();
      offExit();
    };
  }, [session]);

  useEffect(() => {
    setProviders(workspace.providers);
    if (!providerId && workspace.providers[0]) setProviderId(workspace.providers[0].id);
  }, [workspace.providers]);

  useEffect(() => {
    if (!workspace.path) return;
    void refreshSessions();
    void listAgentProviders().then((items) => {
      setProviders(items);
      if (!items.some((item) => item.id === providerId) && items[0]) setProviderId(items[0].id);
    });
  }, [workspace.path]);

  async function refreshSessions() {
    if (!workspace.path) return;
    const items = await listAgentSessions(workspace.path);
    setSessions(items);
  }

  async function runProvider() {
    if (!workspace.path || !selectedProvider) return;
    setOutput([]);
    const next = await startAgentSession({
      root: workspace.path,
      providerId,
      taskId,
      mode,
      prompt,
      context: selectedTask ? selectedTask.prompt : "Workspace: " + workspace.name,
      acceptanceCriteria: mode === "planner" ? "Produce a plan that can be approved before coding." : "Report changed files and verification results.",
      constraints: "Use provider-neutral instructions. Do not assume a specific vendor runtime.",
      allowedFiles: selectedTask?.worktree ? [selectedTask.worktree] : [],
      customCommand,
      rows: 24,
      cols: 100,
    });
    if (next) {
      setSession(next);
      setPrompt("");
      await refreshSessions();
    }
  }

  async function stopSession() {
    if (!session || !workspace.path) return;
    await stopAgentSession(workspace.path, session.id);
    setSession({ ...session, status: "stopped" });
    await refreshSessions();
  }

  async function sendInput(event: React.FormEvent) {
    event.preventDefault();
    if (!session || session.status !== "running" || !input) return;
    await sendAgentInput({ root: workspace.path, sessionId: session.id, input });
    setInput("");
  }

  async function resizeSession() {
    if (!session || session.status !== "running") return;
    await resizeNativeTerminal({ sessionId: session.terminalId, rows: 32, cols: 120 });
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
          <div className="agent-launcher">
            <label>Provider
              <select value={providerId} onChange={(event) => setProviderId(event.target.value)}>
                {providers.map((provider) => (
                  <option key={provider.id} value={provider.id}>{provider.name}</option>
                ))}
              </select>
            </label>
            <label>Mode
              <select value={mode} onChange={(event) => setMode(event.target.value as AgentMode)}>
                {agentModes.map((item) => <option key={item} value={item}>{titleCase(item)}</option>)}
              </select>
            </label>
            <label>Task
              <select value={taskId} onChange={(event) => setTaskId(event.target.value)}>
                <option value="">Workspace</option>
                {workspace.tasks.map((task) => <option key={task.id} value={task.id}>{task.title}</option>)}
              </select>
            </label>
            {selectedProvider?.id === "custom-local" ? (
              <label>Command
                <input value={customCommand} onChange={(event) => setCustomCommand(event.target.value)} placeholder="agent command" />
              </label>
            ) : null}
            <textarea value={prompt} onChange={(event) => setPrompt(event.target.value)} placeholder="Task, context, acceptance criteria, constraints, allowed files, previous plan, expected output..." />
          </div>
          {selectedProvider && selectedProvider.status !== "installed" && selectedProvider.id !== "custom-local" ? (
            <p className="diagnostic">Provider not installed. Install command: {selectedProvider.setupHint}</p>
          ) : null}
          <div className="provider-actions">
            <button disabled={!canStart} onClick={runProvider}>
              <Play size={15} /> Start Session
            </button>
            <button disabled={!session || session.status !== "running"} onClick={stopSession}>
              <Square size={15} /> Stop
            </button>
            <button disabled={!session || session.status !== "running"} onClick={resizeSession}>
              <Maximize2 size={15} /> 120x32
            </button>
          </div>
          {session ? (
            <div className="terminal-session-meta">
              <span>{session.status}</span>
              <span>{session.providerName}</span>
              <span>{titleCase(session.mode)}</span>
              <code>{session.transcriptPath}</code>
            </div>
          ) : null}
          <pre className="terminal-output">{output.length ? output.join("") : "Start a provider session to stream native terminal output here. No auth token is injected by Thanos."}</pre>
          <form className="terminal-input" onSubmit={sendInput}>
            <input value={input} onChange={(event) => setInput(event.target.value)} placeholder="Send input to the running PTY" />
            <button disabled={!session || session.status !== "running" || !input}><Send size={15} /> Send</button>
          </form>
        </div>
      </Panel>
      <Panel title="Sessions" meta="persisted">
        <div className="agent-session-list">
          {sessions.length === 0 ? <p className="empty">No agent sessions persisted for this workspace.</p> : null}
          {sessions.map((item) => (
            <button key={item.id} onClick={() => setSession(item)} type="button">
              <Bot size={15} />
              <span>{item.providerName} · {titleCase(item.mode)} · {item.status}</span>
              <code>{item.transcriptPath}</code>
            </button>
          ))}
        </div>
      </Panel>
    </div>
  );
}

function titleCase(value: string) {
  return value.split(/[_-]/g).filter(Boolean).map((part) => part.slice(0, 1).toUpperCase() + part.slice(1)).join(" ");
}
