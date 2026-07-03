import { Package, Plus } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { Task, TaskStatus } from "../../domain/models";
import { TaskCard } from "./TaskCard";

export function BoardColumn({
  id,
  title,
  icon: Icon,
  tasks,
  selectedTaskId,
  onSelectTask,
  onAddTask,
  renderSortableTask,
}: {
  id: TaskStatus;
  title: string;
  icon: LucideIcon;
  tasks: Task[];
  selectedTaskId: string;
  onSelectTask: (taskId: string) => void;
  onAddTask?: () => void;
  renderSortableTask: (task: Task, active: boolean, onSelectTask: (taskId: string) => void) => React.ReactNode;
}) {
  return (
    <section className="flex min-h-0 w-72 shrink-0 flex-col rounded-xl border border-slate-800 bg-slate-900/60">
      <header className="flex items-center justify-between border-b border-slate-800 p-3">
        <span className="inline-flex items-center gap-2 text-sm font-medium">
          <Icon size={16} className="text-text-muted" />
          {title}
        </span>
        <span className="rounded-lg bg-slate-800 px-2 py-0.5 text-xs text-text-muted">{tasks.length}</span>
      </header>
      <div className="grid min-h-0 flex-1 content-start gap-3 overflow-y-auto p-3" data-column-id={id}>
        {tasks.length === 0 ? (
          <div className="grid place-items-center gap-1 rounded-xl border border-dashed border-slate-700 bg-slate-950/40 px-3 py-8 text-center">
            <Package size={20} className="text-slate-600" />
            <p className="text-sm font-medium text-text-muted">No tasks</p>
            <p className="text-xs text-slate-600">Drag tasks here or create new</p>
          </div>
        ) : (
          tasks.map((task) => renderSortableTask(task, selectedTaskId === task.id, onSelectTask))
        )}
      </div>
      <button
        onClick={onAddTask}
        className="flex items-center gap-2 border-t border-slate-800 px-3 py-2.5 text-sm text-text-muted transition hover:bg-slate-800/50 hover:text-text-main"
      >
        <Plus size={15} /> Add task
      </button>
    </section>
  );
}
