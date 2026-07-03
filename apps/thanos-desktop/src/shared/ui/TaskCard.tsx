import { Bot, Pencil } from "lucide-react";
import type { DraggableAttributes } from "@dnd-kit/core";
import type { SyntheticListenerMap } from "@dnd-kit/core/dist/hooks/utilities";
import type { Task } from "../../domain/models";
import { AgentAvatar } from "./AgentAvatar";
import { PriorityBadge } from "./PriorityBadge";

export function TaskCard({ task, active, dragAttributes, dragListeners, setNodeRef, style, onEdit }: {
  task: Task;
  active: boolean;
  dragAttributes?: DraggableAttributes;
  dragListeners?: SyntheticListenerMap;
  setNodeRef?: (node: HTMLElement | null) => void;
  style?: React.CSSProperties;
  onEdit?: (taskId: string) => void;
}) {
  return (
    <article
      ref={setNodeRef}
      style={style}
      {...dragAttributes}
      {...dragListeners}
      className={`group relative cursor-grab rounded-xl border bg-bg-card p-3 shadow-lg shadow-black/20 transition hover:border-blue-info/60 hover:bg-slate-900/90 ${active ? "border-purple-primary" : "border-slate-800"}`}
    >
      {onEdit ? (
        <button
          onClick={(event) => {
            event.stopPropagation();
            onEdit(task.id);
          }}
          className="absolute right-2 top-2 hidden rounded-md border border-slate-700 bg-slate-950/90 p-1 text-text-muted hover:border-blue-info hover:text-blue-info group-hover:block"
          aria-label={`Edit ${task.title}`}
        >
          <Pencil size={13} />
        </button>
      ) : null}
      <div className="flex items-center justify-between gap-2">
        <span className="text-xs text-text-muted">{task.id}</span>
        <PriorityBadge priority={task.priority} />
      </div>
      <h3 className="mt-2 text-sm font-semibold text-text-main">{task.title}</h3>
      <p className="mt-1 line-clamp-2 text-xs text-text-muted">{task.description}</p>
      <div className="mt-3 flex flex-wrap gap-1">
        {task.tags.map((tag) => (
          <span key={tag} className="rounded-lg bg-slate-800 px-2 py-1 text-xs text-slate-300">
            {tag}
          </span>
        ))}
      </div>
      <div className="mt-3 h-1.5 overflow-hidden rounded-full bg-slate-800">
        <div className="h-full rounded-full bg-green-success" style={{ width: `${task.progress}%` }} />
      </div>
      <div className="mt-3 flex items-center justify-between gap-2">
        <AgentAvatar label={task.assignedAgent} />
        <Bot size={14} className="text-text-muted" />
      </div>
    </article>
  );
}
