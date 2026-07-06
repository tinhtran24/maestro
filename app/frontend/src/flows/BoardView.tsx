import { Plus, Search } from "lucide-react";
import { useState } from "react";
import type { Task, TaskStatus, Workspace } from "../app/types";
import { createTask, updateTaskFlags, updateTaskStatus } from "../services/wails";
import { StatusBadge } from "../shared/Badge";
import { Panel } from "../shared/Panel";

const columns: Array<{ id: TaskStatus; title: string }> = [
  { id: "backlog", title: "Backlog" },
  { id: "in_progress", title: "In Progress" },
  { id: "waiting", title: "Waiting" },
  { id: "committing", title: "Committing" },
  { id: "done", title: "Done" },
  { id: "failed", title: "Failed" },
  { id: "cancelled", title: "Cancelled" },
];

const transitions: Record<TaskStatus, TaskStatus[]> = {
  backlog: ["in_progress", "cancelled"],
  in_progress: ["waiting", "committing", "failed", "cancelled"],
  waiting: ["in_progress", "committing", "failed", "cancelled"],
  committing: ["done", "failed", "cancelled"],
  done: [],
  failed: ["backlog", "in_progress", "cancelled"],
  cancelled: ["backlog"],
};

export function BoardView({ workspace, onReload }: { workspace: Workspace; onReload: () => Promise<void> }) {
  const [title, setTitle] = useState("");
  const [prompt, setPrompt] = useState("");
  const [query, setQuery] = useState("");
  const [showArchived, setShowArchived] = useState(false);
  const [showDeleted, setShowDeleted] = useState(false);
  const [error, setError] = useState("");

  async function submitTask(event: React.FormEvent) {
    event.preventDefault();
    if (!workspace.path || (!title.trim() && !prompt.trim())) return;
    setError("");
    await createTask({ root: workspace.path, title, prompt, flow: "implement", agent: "auto" });
    setTitle("");
    setPrompt("");
    await onReload();
  }

  async function moveTask(taskId: string, status: TaskStatus) {
    if (!workspace.path) return;
    setError("");
    try {
      await updateTaskStatus({ root: workspace.path, taskId, status });
      await onReload();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : String(reason));
    }
  }

  async function setTaskFlags(task: Task, flags: Partial<Pick<Task, "archived" | "deleted" | "tombstone">>) {
    if (!workspace.path) return;
    setError("");
    try {
      await updateTaskFlags({
        root: workspace.path,
        taskId: task.id,
        archived: flags.archived ?? task.archived,
        deleted: flags.deleted ?? task.deleted,
        tombstone: flags.tombstone ?? task.tombstone,
      });
      await onReload();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : String(reason));
    }
  }

  const visibleTasks = workspace.tasks.filter((task) => {
    if (!showArchived && task.archived) return false;
    if (!showDeleted && (task.deleted || task.tombstone)) return false;
    const text = `${task.id} ${task.title} ${task.prompt} ${task.agent} ${task.flow}`.toLowerCase();
    return !query.trim() || text.includes(query.trim().toLowerCase());
  });

  const blockedTasks = visibleTasks.filter((task) => task.blocked && !task.deleted && !task.tombstone);
  const archivedTasks = workspace.tasks.filter((task) => task.archived);
  const deletedTasks = workspace.tasks.filter((task) => task.deleted || task.tombstone);

  function renderTask(task: Task) {
    return (
      <article key={task.id} className={`task-card ${task.blocked ? "blocked" : ""}`}>
        <div className="task-card-head">
          <strong>{task.title}</strong>
          <StatusBadge status={task.status} />
        </div>
        <p>{task.prompt}</p>
        <div className="task-meta">
          <span>{task.agent}</span>
          <span>{task.flow}</span>
          <span>${task.usageUsd.toFixed(2)}</span>
          {task.dependencies.length ? <span>{task.dependencies.length} deps</span> : null}
          {task.blocked ? <span>blocked</span> : null}
          {task.failureCategory ? <span>{task.failureCategory}</span> : null}
        </div>
        <div className="task-actions">
          {transitions[task.status].map((target) => (
            <button key={target} disabled={task.blocked && target === "in_progress"} onClick={() => moveTask(task.id, target)}>
              {target.replace("_", " ")}
            </button>
          ))}
          <button onClick={() => setTaskFlags(task, { archived: !task.archived })}>{task.archived ? "unarchive" : "archive"}</button>
          <button onClick={() => setTaskFlags(task, { deleted: !task.deleted, tombstone: task.tombstone && task.deleted ? false : task.tombstone })}>{task.deleted ? "restore" : "delete"}</button>
          {!task.tombstone ? <button onClick={() => setTaskFlags(task, { deleted: true, tombstone: true })}>tombstone</button> : null}
        </div>
      </article>
    );
  }

  return (
    <div className="view-grid">
      <section className="hero-strip">
        <div>
          <h2>Task board</h2>
          <p>Coordinate agent work from backlog to review without hiding the run state.</p>
        </div>
        <form className="quick-create" onSubmit={submitTask}>
          <input value={title} onChange={(event) => setTitle(event.target.value)} placeholder="Task title" />
          <input value={prompt} onChange={(event) => setPrompt(event.target.value)} placeholder="Prompt or acceptance criteria" />
          <button disabled={!workspace.path}><Plus size={16} /> New Task</button>
        </form>
      </section>
      <section className="board-tools">
        <label><Search size={15} /><input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search tasks" /></label>
        <button className={showArchived ? "enabled" : ""} onClick={() => setShowArchived((value) => !value)}>Archived {archivedTasks.length}</button>
        <button className={showDeleted ? "enabled" : ""} onClick={() => setShowDeleted((value) => !value)}>Deleted {deletedTasks.length}</button>
        <span>{blockedTasks.length} blocked</span>
      </section>
      {error ? <div className="error-banner">{error}</div> : null}
      <div className="board">
        {columns.map((column) => {
          const tasks = visibleTasks.filter((task) => task.status === column.id);
          return (
            <Panel key={column.id} title={column.title} meta={`${tasks.length}`}>
              <div className="task-list">
                {tasks.map(renderTask)}
                {tasks.length === 0 ? <p className="empty">No tasks here.</p> : null}
              </div>
            </Panel>
          );
        })}
      </div>
    </div>
  );
}
