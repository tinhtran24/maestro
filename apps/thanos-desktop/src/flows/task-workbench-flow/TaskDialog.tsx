import { useEffect, useMemo, useState } from "react";
import type { Priority, Task } from "../../domain/models";
import { useWorkbenchStore } from "../../state/workbenchStore";

export function TaskDialog() {
  const dialog = useWorkbenchStore((state) => state.taskDialog);
  const tasks = useWorkbenchStore((state) => state.tasks);
  const createTask = useWorkbenchStore((state) => state.createTask);
  const editTask = useWorkbenchStore((state) => state.editTask);
  const close = useWorkbenchStore((state) => state.closeTaskDialog);
  const task = useMemo(() => tasks.find((item) => item.id === dialog?.taskId), [dialog?.taskId, tasks]);
  const [draft, setDraft] = useState({ title: "", description: "", priority: "P2" as Priority, assignedAgent: "" });

  useEffect(() => {
    if (!dialog) return;
    setDraft({
      title: task?.title ?? "",
      description: task?.description ?? "",
      priority: task?.priority ?? "P2",
      assignedAgent: task?.assignedAgent ?? "",
    });
  }, [dialog, task]);

  if (!dialog) return null;

  function submit() {
    const input = {
      title: draft.title.trim() || "Untitled task",
      description: draft.description.trim(),
      priority: draft.priority,
      assignedAgent: draft.assignedAgent.trim(),
    };
    if (dialog?.mode === "edit" && task) editTask(task.id, input);
    else createTask(input);
  }

  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-black/60 p-4">
      <div className="w-full max-w-xl rounded-lg border border-slate-800 bg-bg-card p-5 shadow-2xl shadow-black/40">
        <header className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">{dialog.mode === "edit" ? "Edit Task" : "Add Task"}</h2>
          <button onClick={close} className="rounded-md border border-slate-700 px-2 py-1 text-sm text-text-muted hover:border-slate-500 hover:text-text-main">Close</button>
        </header>
        <div className="mt-4 grid gap-3">
          <Field label="Title" value={draft.title} onChange={(title) => setDraft((current) => ({ ...current, title }))} />
          <label className="grid gap-1 text-sm">
            <span className="text-text-muted">Description</span>
            <textarea value={draft.description} onChange={(event) => setDraft((current) => ({ ...current, description: event.target.value }))} className="min-h-24 rounded-lg border border-slate-800 bg-slate-950/70 px-3 py-2 text-text-main" />
          </label>
          <div className="grid grid-cols-2 gap-3">
            <label className="grid gap-1 text-sm">
              <span className="text-text-muted">Priority</span>
              <select value={draft.priority} onChange={(event) => setDraft((current) => ({ ...current, priority: event.target.value as Priority }))} className="h-9 rounded-lg border border-slate-800 bg-slate-950/70 px-3 text-text-main">
                {(["P0", "P1", "P2", "P3"] satisfies Priority[]).map((priority) => <option key={priority}>{priority}</option>)}
              </select>
            </label>
            <Field label="Assigned agent" value={draft.assignedAgent} onChange={(assignedAgent) => setDraft((current) => ({ ...current, assignedAgent }))} />
          </div>
        </div>
        <footer className="mt-5 flex justify-end gap-2">
          <button onClick={close} className="rounded-lg border border-slate-800 px-3 py-2 text-sm text-text-muted hover:border-slate-600 hover:text-text-main">Cancel</button>
          <button onClick={submit} className="rounded-lg bg-purple-primary px-3 py-2 text-sm font-medium text-white hover:bg-purple-hover">{dialog.mode === "edit" ? "Save Task" : "Add Task"}</button>
        </footer>
      </div>
    </div>
  );
}

function Field({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
  return (
    <label className="grid gap-1 text-sm">
      <span className="text-text-muted">{label}</span>
      <input value={value} onChange={(event) => onChange(event.target.value)} className="h-9 rounded-lg border border-slate-800 bg-slate-950/70 px-3 text-text-main" />
    </label>
  );
}
