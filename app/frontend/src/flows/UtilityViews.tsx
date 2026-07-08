import { CalendarClock, File, Folder, GitCompare, MessageSquare, PenTool, RefreshCw, Save, Settings } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import type { FileTreeEntry, TaskDiffPreviewInfo, Workspace, WorkspaceFileInfo } from "../app/types";
import {
  listWorkspaceFiles,
  previewTaskDiff,
  readWorkspaceFile,
  runRoutineScheduler,
  saveAutomation,
  triggerRoutine,
  upsertRoutine,
  writeWorkspaceFile,
} from "../services/wails";
import { Panel } from "../shared/Panel";

type AutomationToggle = "autoImplement" | "autoTest" | "autoSubmit" | "autoRetry";

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

  async function runScheduler() {
    if (!workspace.path) return;
    await runRoutineScheduler({ root: workspace.path });
    await onReload();
  }

  async function runRoutine(routineId: string) {
    if (!workspace.path) return;
    await triggerRoutine({ root: workspace.path, routineId });
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
            <span>runs {routine.runCount ?? 0} · failures {routine.failureCount ?? 0}{routine.nextRunAt ? ` · next ${routine.nextRunAt}` : ""}</span>
            <p>{routine.prompt}</p>
            {routine.disabledReason ? <p className="diagnostic">Stopped: {routine.disabledReason}</p> : null}
            <button onClick={() => runRoutine(routine.id)} disabled={!workspace.path || !routine.enabled} type="button">Trigger Routine</button>
          </article>
        ))}
      </div>
      <button className="wide-action" onClick={runScheduler} disabled={!workspace.path} type="button"><CalendarClock size={16} /> Run Scheduler Tick</button>
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

export function FileExplorerView({ workspace }: { workspace: Workspace }) {
  const [tree, setTree] = useState<FileTreeEntry[]>([]);
  const [selected, setSelected] = useState<WorkspaceFileInfo | null>(null);
  const [content, setContent] = useState("");
  const [status, setStatus] = useState("");
  const [diffTaskId, setDiffTaskId] = useState(workspace.tasks[0]?.id ?? "");
  const [diff, setDiff] = useState<TaskDiffPreviewInfo | null>(null);
  const canSave = Boolean(selected && !selected.readOnly && selected.content !== content);
  const taskOptions = useMemo(() => workspace.tasks.filter((task) => task.worktree), [workspace.tasks]);

  async function loadTree() {
    if (!workspace.path) return;
    setStatus("Loading Files");
    try {
      const next = await listWorkspaceFiles({ root: workspace.path, maxDepth: 3 });
      setTree(next.entries);
      setStatus("");
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Unable To Load Files");
    }
  }

  async function openFile(path: string) {
    if (!workspace.path) return;
    setStatus("Opening File");
    try {
      const file = await readWorkspaceFile({ root: workspace.path, path });
      if (file) {
        setSelected(file);
        setContent(file.content);
      }
      setStatus("");
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Unable To Open File");
    }
  }

  async function saveFile() {
    if (!workspace.path || !selected || selected.readOnly) return;
    setStatus("Saving File");
    try {
      const saved = await writeWorkspaceFile({ root: workspace.path, path: selected.path, content });
      if (saved) {
        setSelected(saved);
        setContent(saved.content);
      }
      await loadTree();
      setStatus("Saved");
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Unable To Save File");
    }
  }

  async function loadDiff(taskId = diffTaskId) {
    if (!workspace.path || !taskId) return;
    setStatus("Loading Diff");
    try {
      const next = await previewTaskDiff({ root: workspace.path, taskId });
      setDiff(next);
      setStatus("");
    } catch (error) {
      setDiff(null);
      setStatus(error instanceof Error ? error.message : "Unable To Load Diff");
    }
  }

  useEffect(() => {
    void loadTree();
  }, [workspace.path]);

  useEffect(() => {
    setDiffTaskId(workspace.tasks[0]?.id ?? "");
  }, [workspace.tasks]);

  return (
    <div className="file-explorer">
      <Panel title="Files" meta={workspace.path || "No Workspace"}>
        <div className="file-toolbar">
          <button onClick={loadTree} disabled={!workspace.path} type="button"><RefreshCw size={15} /> Refresh</button>
          {status ? <span>{status}</span> : null}
        </div>
        <div className="file-tree">
          {tree.length === 0 ? <p className="empty">No files loaded.</p> : null}
          {tree.map((entry) => (
            <FileTreeNode key={entry.path} entry={entry} onOpen={openFile} />
          ))}
        </div>
      </Panel>
      <Panel title={selected?.name ?? "File Preview"} meta={selected?.path ?? "Select File"}>
        <div className="file-editor">
          <textarea value={content} onChange={(event) => setContent(event.target.value)} readOnly={!selected || selected.readOnly} />
          <div className="file-editor-actions">
            <span>{selected ? `${selected.encoding} · ${selected.size} bytes${selected.readOnly ? " · Read Only" : ""}` : "No file selected"}</span>
            <button onClick={saveFile} disabled={!canSave} type="button"><Save size={15} /> Save File</button>
          </div>
        </div>
      </Panel>
      <Panel title="Task Diff" meta="worktree preview">
        <div className="file-diff-tools">
          <select value={diffTaskId} onChange={(event) => setDiffTaskId(event.target.value)}>
            {taskOptions.map((task) => (
              <option key={task.id} value={task.id}>{task.title || task.id}</option>
            ))}
          </select>
          <button onClick={() => loadDiff()} disabled={!workspace.path || !diffTaskId} type="button"><GitCompare size={15} /> Preview Diff</button>
        </div>
        <pre className="file-diff">{diff ? `${diff.diffStat}\n\n${diff.diff}`.trim() || "No diff." : "Select a task worktree."}</pre>
      </Panel>
    </div>
  );
}

function FileTreeNode({ entry, onOpen }: { entry: FileTreeEntry; onOpen: (path: string) => void }) {
  const isDirectory = entry.kind === "directory";
  const Icon = isDirectory ? Folder : File;
  return (
    <div className="file-tree-node">
      <button onClick={() => !isDirectory && onOpen(entry.path)} disabled={isDirectory} type="button">
        <Icon size={15} />
        <span>{entry.name}</span>
      </button>
      {entry.children?.length ? (
        <div className="file-tree-children">
          {entry.children.map((child) => (
            <FileTreeNode key={child.path} entry={child} onOpen={onOpen} />
          ))}
        </div>
      ) : null}
    </div>
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
  const [maxConcurrent, setMaxConcurrent] = useState(String(workspace.automation.maxConcurrentRoutineTasks || 3));
  const [failureLimit, setFailureLimit] = useState(String(workspace.automation.circuitBreakerFailureLimit || 3));
  async function toggle(key: AutomationToggle) {
    if (!workspace.path) return;
    await saveAutomation({ root: workspace.path, automation: { ...workspace.automation, [key]: !workspace.automation[key] } });
    await onReload();
  }
  async function saveLimits() {
    if (!workspace.path) return;
    await saveAutomation({
      root: workspace.path,
      automation: {
        ...workspace.automation,
        maxConcurrentRoutineTasks: Number(maxConcurrent) || 3,
        circuitBreakerFailureLimit: Number(failureLimit) || 3,
      },
    });
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
        <div className="automation-limits">
          <label>Max Concurrent Routine Tasks<input value={maxConcurrent} onChange={(event) => setMaxConcurrent(event.target.value)} type="number" min="1" /></label>
          <label>Circuit Breaker Failure Limit<input value={failureLimit} onChange={(event) => setFailureLimit(event.target.value)} type="number" min="1" /></label>
          <button onClick={saveLimits} type="button">Save Limits</button>
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
