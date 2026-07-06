import type { TaskStatus } from "../app/types";

const tone: Record<TaskStatus, string> = {
  backlog: "border-slate-700 text-slate-300",
  running: "border-sky-500/40 text-sky-300",
  waiting: "border-amber-500/40 text-amber-300",
  review: "border-orange-500/40 text-orange-300",
  done: "border-emerald-500/40 text-emerald-300",
  failed: "border-red-500/40 text-red-300",
};

export function StatusBadge({ status }: { status: TaskStatus }) {
  return (
    <span className={`inline-flex h-6 items-center rounded-md border px-2 text-xs font-medium ${tone[status]}`}>
      {status.replace("_", " ")}
    </span>
  );
}
