import { ChevronDown, Grid2X2, List, MoreVertical, Plus, Search, X } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import type { Task, TaskStatus, Workspace } from "../app/types";
import { createTask, startTaskTurn, updateTaskFlags, updateTaskStatus } from "../services/wails";
import { taskMoveActionForStatus } from "./boardActions";

const columns: Array<{ id: TaskStatus; title: string; tone: string }> = [
  { id: "backlog", title: "Backlog", tone: "blue" },
  { id: "in_progress", title: "In Progress", tone: "sky" },
  { id: "waiting", title: "Waiting", tone: "amber" },
  { id: "committing", title: "Committing", tone: "violet" },
  { id: "done", title: "Done", tone: "emerald" },
  { id: "failed", title: "Failed", tone: "red" },
  { id: "cancelled", title: "Cancelled", tone: "slate" },
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

const priorityOptions = ["All", "Low", "Medium", "High"];

function statusLabel(status: TaskStatus) {
  return columns.find((column) => column.id === status)?.title ?? status.replace("_", " ");
}

function taskPriority(task: Task) {
  if (task.failureCategory || task.blocked || task.status === "failed") return "High";
  if (task.status === "done" || task.status === "cancelled") return "Low";
  return "Medium";
}

function taskLabel(task: Task) {
  if (task.failureCategory) return task.failureCategory;
  if (task.blocked) return "Blocked";
  if (task.flow && task.flow !== "implement") return task.flow;
  if (task.lastTestResult?.status) return task.lastTestResult.status;
  return task.agent || "Task";
}

function taskInitial(task: Task) {
  return (task.agent || task.flow || task.title || "T").slice(0, 1).toUpperCase();
}

export function BoardView({ workspace, onReload }: { workspace: Workspace; onReload: () => Promise<void> }) {
  const [title, setTitle] = useState("");
  const [prompt, setPrompt] = useState("");
  const [relatedTaskId, setRelatedTaskId] = useState("");
  const [query, setQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<TaskStatus | "all">("all");
  const [labelFilter, setLabelFilter] = useState("all");
  const [priorityFilter, setPriorityFilter] = useState("All");
  const [selectedStatus, setSelectedStatus] = useState<TaskStatus>("backlog");
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [viewMode, setViewMode] = useState<"board" | "list">("board");
  const [showArchived, setShowArchived] = useState(false);
  const [showDeleted, setShowDeleted] = useState(false);
  const [error, setError] = useState("");
  const searchRef = useRef<HTMLInputElement>(null);
  const titleRef = useRef<HTMLInputElement>(null);

  const labels = useMemo(() => {
    return Array.from(new Set(workspace.tasks.map(taskLabel))).sort((a, b) => a.localeCompare(b));
  }, [workspace.tasks]);

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      const target = event.target as HTMLElement | null;
      const isTyping = target?.tagName === "INPUT" || target?.tagName === "TEXTAREA" || target?.tagName === "SELECT";
      if (event.key === "Escape" && drawerOpen) {
        setDrawerOpen(false);
      }
      if (event.key.toLowerCase() === "n" && !isTyping) {
        event.preventDefault();
        openDrawer("backlog");
      }
      if (event.key === "/" && !isTyping) {
        event.preventDefault();
        searchRef.current?.focus();
      }
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [drawerOpen]);

  useEffect(() => {
    if (drawerOpen) {
      window.setTimeout(() => titleRef.current?.focus(), 0);
    }
  }, [drawerOpen]);

  function openDrawer(status: TaskStatus) {
    setSelectedStatus(status);
    setError("");
    setDrawerOpen(true);
  }

  function closeDrawer() {
    setDrawerOpen(false);
    setTitle("");
    setPrompt("");
    setRelatedTaskId("");
  }

  async function submitTask(event?: { preventDefault: () => void }) {
    event?.preventDefault();
    if (!workspace.path || !title.trim() || !prompt.trim()) return;
    setError("");
    await createTask({
      root: workspace.path,
      title,
      prompt,
      flow: "implement",
      agent: "auto",
      status: selectedStatus,
      dependencies: relatedTaskId ? [relatedTaskId] : undefined,
    });
    closeDrawer();
    await onReload();
  }

  async function moveTask(taskId: string, status: TaskStatus) {
    if (!workspace.path) return;
    setError("");
    try {
      if (taskMoveActionForStatus(status) === "start-turn") {
        await startTaskTurn({ root: workspace.path, taskId, step: "Implementation" });
      } else {
        await updateTaskStatus({ root: workspace.path, taskId, status });
      }
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
    if (statusFilter !== "all" && task.status !== statusFilter) return false;
    if (labelFilter !== "all" && taskLabel(task) !== labelFilter) return false;
    if (priorityFilter !== "All" && taskPriority(task) !== priorityFilter) return false;
    const text = `${task.id} ${task.title} ${task.prompt} ${task.agent} ${task.flow}`.toLowerCase();
    return !query.trim() || text.includes(query.trim().toLowerCase());
  });

  const canCreate = Boolean(workspace.path && title.trim() && prompt.trim());

  function renderTask(task: Task) {
    const priority = taskPriority(task);
    const label = taskLabel(task);
    return (
      <article key={task.id} className={`board-task-card ${task.blocked ? "is-blocked" : ""}`}>
        <div className="board-task-title">
          <strong>{task.title}</strong>
          <button aria-label={`Task Actions For ${task.title}`} type="button">
            <MoreVertical size={15} />
          </button>
        </div>
        <p>{task.prompt || "No Description"}</p>
        <div className="board-task-labels">
          <span className={`task-chip ${task.status}`}>{label}</span>
        </div>
        <div className="board-task-meta">
          <span>#{task.id.replace(/^task-/, "").slice(0, 4)}</span>
          <span>{priority}</span>
          <span className="task-avatar">{taskInitial(task)}</span>
        </div>
        <div className="board-task-actions">
          {transitions[task.status].map((target) => (
            <button key={target} disabled={task.blocked && target === "in_progress"} onClick={() => moveTask(task.id, target)} type="button">
              {statusLabel(target)}
            </button>
          ))}
          <button onClick={() => setTaskFlags(task, { archived: !task.archived })} type="button">
            {task.archived ? "Unarchive" : "Archive"}
          </button>
          <button onClick={() => setTaskFlags(task, { deleted: !task.deleted, tombstone: task.tombstone && task.deleted ? false : task.tombstone })} type="button">
            {task.deleted ? "Restore" : "Delete"}
          </button>
        </div>
      </article>
    );
  }

  return (
    <div className={`board-screen ${drawerOpen ? "has-drawer" : ""}`}>
      <section className="board-workspace">
        <div className="board-tools">
          <label className="board-search">
            <Search size={17} />
            <input ref={searchRef} value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search Tasks..." />
          </label>
          <label className="board-select">
            <span>Status: </span>
            <select value={statusFilter} onChange={(event) => setStatusFilter(event.target.value as TaskStatus | "all")}>
              <option value="all">All</option>
              {columns.map((column) => (
                <option key={column.id} value={column.id}>{column.title}</option>
              ))}
            </select>
          </label>
          <label className="board-select">
            <span>Label: </span>
            <select value={labelFilter} onChange={(event) => setLabelFilter(event.target.value)}>
              <option value="all">All</option>
              {labels.map((label) => (
                <option key={label} value={label}>{label}</option>
              ))}
            </select>
          </label>
          <label className="board-select">
            <span>Priority: </span>
            <select value={priorityFilter} onChange={(event) => setPriorityFilter(event.target.value)}>
              {priorityOptions.map((priority) => (
                <option key={priority} value={priority}>{priority}</option>
              ))}
            </select>
          </label>
          <button className={showArchived ? "enabled" : ""} onClick={() => setShowArchived((value) => !value)} type="button">Archived</button>
          <button className={showDeleted ? "enabled" : ""} onClick={() => setShowDeleted((value) => !value)} type="button">Deleted</button>
          <label className="board-select board-group">
            <span>Group By: </span>
            <select value="status" onChange={() => undefined}>
              <option value="status">Status</option>
            </select>
          </label>
          <div className="board-view-toggle" aria-label="Board View Toggle">
            <button className={viewMode === "board" ? "active" : ""} onClick={() => setViewMode("board")} aria-label="Board View" type="button">
              <Grid2X2 size={18} />
            </button>
            <button className={viewMode === "list" ? "active" : ""} onClick={() => setViewMode("list")} aria-label="List View" type="button">
              <List size={18} />
            </button>
          </div>
        </div>

        {error ? <div className="error-banner">{error}</div> : null}

        <div className={viewMode === "board" ? "board-kanban" : "board-kanban list-mode"}>
          {columns.map((column) => {
            const tasks = visibleTasks.filter((task) => task.status === column.id);
            return (
              <section key={column.id} className={`board-column ${column.tone}`}>
                <header className="board-column-header">
                  <div>
                    <span className="status-dot" />
                    <h2>{column.title}</h2>
                  </div>
                  <span className="column-count">{tasks.length}</span>
                </header>
                <div className="task-list">
                  {tasks.map(renderTask)}
                  {tasks.length === 0 ? (
                    <div className="board-empty">
                      <strong>No Tasks Here</strong>
                      <span>{column.title} tasks will appear here.</span>
                    </div>
                  ) : null}
                </div>
                <button className="add-task-link" onClick={() => openDrawer(column.id)} type="button">
                  <Plus size={16} /> Add Task
                </button>
              </section>
            );
          })}
        </div>
      </section>

      {drawerOpen ? (
        <aside className="new-task-drawer" aria-label="New Task">
          <form onSubmit={submitTask} onKeyDown={(event) => {
            if (event.ctrlKey && event.key === "Enter") {
              void submitTask(event);
            }
          }}>
            <header>
              <h2>New Task</h2>
              <button aria-label="Close New Task" onClick={closeDrawer} type="button">
                <X size={22} />
              </button>
            </header>
            <div className="drawer-body">
              <label>
                <span>Title <em>*</em></span>
                <input ref={titleRef} value={title} onChange={(event) => setTitle(event.target.value)} placeholder="Enter a clear, concise title..." required />
              </label>
              <label>
                <span>Prompt / Description <em>*</em></span>
                <textarea value={prompt} onChange={(event) => setPrompt(event.target.value)} placeholder="Describe what needs to be done and how we'll know it's complete..." required />
              </label>
              <label>
                <span>Related Task</span>
                <div className="related-select">
                  <select value={relatedTaskId} onChange={(event) => setRelatedTaskId(event.target.value)}>
                    <option value="">Search Tasks...</option>
                    {workspace.tasks.map((task) => (
                      <option key={task.id} value={task.id}>{task.title}</option>
                    ))}
                  </select>
                  <ChevronDown size={18} />
                </div>
                <small>Link an existing task that this is related to.</small>
              </label>
            </div>
            <footer>
              <button className="secondary" onClick={closeDrawer} type="button">Cancel</button>
              <button className="primary" disabled={!canCreate} type="submit">Create Task</button>
            </footer>
          </form>
        </aside>
      ) : null}
    </div>
  );
}
