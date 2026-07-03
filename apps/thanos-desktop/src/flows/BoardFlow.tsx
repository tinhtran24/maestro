import { DndContext, type DragEndEvent, PointerSensor, useDroppable, useSensor, useSensors } from "@dnd-kit/core";
import { SortableContext, useSortable, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { CheckCircle2, GitPullRequest, HelpCircle, Inbox, ListFilter, ListTodo, PlayCircle, Search, ShieldCheck } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { Task, TaskStatus } from "../domain/models";
import { BoardColumn } from "../shared/ui/BoardColumn";
import { TaskCard } from "../shared/ui/TaskCard";
import { useWorkbenchStore } from "../state/workbenchStore";

const columns: Array<{ id: TaskStatus; title: string; icon: LucideIcon }> = [
  { id: "backlog", title: "Backlog", icon: Inbox },
  { id: "planning", title: "Planning", icon: ListTodo },
  { id: "waiting_approval", title: "Waiting Approval", icon: ShieldCheck },
  { id: "running", title: "In Progress", icon: PlayCircle },
  { id: "in_review", title: "In Review", icon: GitPullRequest },
  { id: "waiting_user", title: "Waiting User", icon: HelpCircle },
  { id: "done", title: "Done", icon: CheckCircle2 },
];

export function BoardFlow() {
  const tasks = useWorkbenchStore((state) => state.tasks);
  const selectedTaskId = useWorkbenchStore((state) => state.selectedTaskId);
  const filter = useWorkbenchStore((state) => state.boardFilter);
  const setFilter = useWorkbenchStore((state) => state.setBoardFilter);
  const selectTask = useWorkbenchStore((state) => state.selectTask);
  const openEditTask = useWorkbenchStore((state) => state.openEditTask);
  const openCreateTask = useWorkbenchStore((state) => state.openCreateTask);
  const moveTask = useWorkbenchStore((state) => state.moveTask);
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 6 } }));
  const visible = filter.trim()
    ? tasks.filter((task) => `${task.id} ${task.title} ${task.description} ${task.tags.join(" ")}`.toLowerCase().includes(filter.toLowerCase()))
    : tasks;

  function onDragEnd(event: DragEndEvent) {
    const taskId = String(event.active.id);
    const status = event.over?.id ? String(event.over.id) as TaskStatus : null;
    if (status) moveTask(taskId, status);
  }

  return (
    <section className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-3 p-4">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-xl font-semibold">Board</h1>
          <p className="mt-0.5 text-sm text-text-muted">{tasks.length ? "Track AI workflow tasks across planning, coding, review and testing." : "No tasks yet. Use New Task to add the first item."}</p>
        </div>
        <div className="flex items-center gap-2">
          <div className="relative">
            <Search size={15} className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-text-muted" />
            <input
              value={filter}
              onChange={(event) => setFilter(event.target.value)}
              className="h-9 w-56 rounded-lg border border-slate-800 bg-slate-900/80 pl-9 pr-3 text-sm text-text-main placeholder:text-text-muted focus:border-slate-600 focus:outline-none"
              placeholder="Search tasks..."
            />
          </div>
          <button className="inline-flex h-9 items-center gap-2 rounded-lg border border-slate-800 bg-slate-900/80 px-3 text-sm text-text-muted hover:border-slate-700 hover:text-text-main">
            <ListFilter size={15} /> Filters
          </button>
        </div>
      </div>
      <DndContext sensors={sensors} onDragEnd={onDragEnd}>
        <div className="flex min-h-0 gap-3 overflow-x-auto pb-2">
          {columns.map((column) => {
            const columnTasks = visible.filter((task) => task.status === column.id);
            return (
              <DroppableColumn key={column.id} id={column.id}>
                <SortableContext items={columnTasks.map((task) => task.id)} strategy={verticalListSortingStrategy}>
                  <BoardColumn
                    {...column}
                    tasks={columnTasks}
                    selectedTaskId={selectedTaskId}
                    onSelectTask={selectTask}
                    onAddTask={openCreateTask}
                    renderSortableTask={(task, active, onSelect) => <SortableTask key={task.id} task={task} active={active} onSelect={onSelect} onEdit={openEditTask} />}
                  />
                </SortableContext>
              </DroppableColumn>
            );
          })}
        </div>
      </DndContext>
    </section>
  );
}

function DroppableColumn({ id, children }: { id: TaskStatus; children: React.ReactNode }) {
  const { setNodeRef } = useDroppable({ id });
  return <div ref={setNodeRef}>{children}</div>;
}

function SortableTask({ task, active, onSelect, onEdit }: { task: Task; active: boolean; onSelect: (taskId: string) => void; onEdit: (taskId: string) => void }) {
  const { attributes, listeners, setNodeRef, transform, transition } = useSortable({ id: task.id });
  return (
    <button className="text-left" onClick={() => onSelect(task.id)}>
      <TaskCard
        task={task}
        active={active}
        setNodeRef={setNodeRef}
        dragAttributes={attributes}
        dragListeners={listeners}
        onEdit={onEdit}
        style={{ transform: CSS.Transform.toString(transform), transition }}
      />
    </button>
  );
}
