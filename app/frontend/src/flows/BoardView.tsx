import { Plus } from "lucide-react";
import { useState } from "react";
import type { TaskStatus, Workspace } from "../app/types";
import { createTask, updateTaskStatus } from "../services/wails";
import { StatusBadge } from "../shared/Badge";
import { Panel } from "../shared/Panel";

const columns: Array<{ id: TaskStatus; title: string }> = [
  { id: "backlog", title: "Backlog" },
  { id: "running", title: "Running" },
  { id: "waiting", title: "Waiting" },
  { id: "review", title: "Review" },
  { id: "done", title: "Done" },
];

export function BoardView({ workspace, onReload }: { workspace: Workspace; onReload: () => Promise<void> }) {
  const [title, setTitle] = useState("");
  const [prompt, setPrompt] = useState("");

  async function submitTask(event: React.FormEvent) {
    event.preventDefault();
    if (!workspace.path || (!title.trim() && !prompt.trim())) return;
    await createTask({ root: workspace.path, title, prompt, flow: "implement", agent: "auto" });
    setTitle("");
    setPrompt("");
    await onReload();
  }

  async function moveTask(taskId: string, status: TaskStatus) {
    if (!workspace.path) return;
    await updateTaskStatus({ root: workspace.path, taskId, status });
    await onReload();
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
      <div className="board">
        {columns.map((column) => {
          const tasks = workspace.tasks.filter((task) => task.status === column.id);
          return (
            <Panel key={column.id} title={column.title} meta={`${tasks.length}`}>
              <div className="task-list">
                {tasks.map((task) => (
                  <article key={task.id} className="task-card">
                    <div className="task-card-head">
                      <strong>{task.title}</strong>
                      <StatusBadge status={task.status} />
                    </div>
                    <p>{task.prompt}</p>
                    <div className="task-meta">
                      <span>{task.agent}</span>
                      <span>{task.flow}</span>
                      <span>${task.usageUsd.toFixed(2)}</span>
                    </div>
                    <div className="task-actions">
                      {columns.map((target) => (
                        target.id !== task.status ? <button key={target.id} onClick={() => moveTask(task.id, target.id)}>{target.title}</button> : null
                      ))}
                    </div>
                  </article>
                ))}
                {tasks.length === 0 ? <p className="empty">No tasks here.</p> : null}
              </div>
            </Panel>
          );
        })}
      </div>
    </div>
  );
}
