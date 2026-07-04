import { CheckCircle2 } from "lucide-react";
import type { CapturedAttachment, ExtractedTask } from "../../features/tasks/quickCapture";

// Step 4 — Create. Final summary before the task is added to the board.
export function CreateTaskFlow({ draft, attachments }: { draft: ExtractedTask; attachments: CapturedAttachment[] }) {
  return (
    <div className="grid gap-4">
      <p className="inline-flex items-center gap-2 rounded-lg border border-green-success/30 bg-green-success/10 p-3 text-sm text-green-success">
        <CheckCircle2 size={16} /> Everything is ready. Create the task to add it to the board.
      </p>
      <dl className="grid gap-2 rounded-lg border border-slate-800 bg-slate-950/40 p-4 text-sm">
        <Item label="Title" value={draft.title.value} />
        <Item label="Priority" value={draft.priority.value} />
        <Item label="Feature" value={draft.feature.value} />
        <Item label="Labels" value={draft.labels.value.join(", ") || "—"} />
        <Item label="Acceptance criteria" value={`${draft.acceptanceCriteria.length} item${draft.acceptanceCriteria.length === 1 ? "" : "s"}`} />
        <Item label="Attachments" value={`${attachments.length}`} />
      </dl>
      <p className="text-xs text-text-muted">The task starts in Backlog. Planning does not start automatically — open the task and start the planner when ready.</p>
    </div>
  );
}

function Item({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <dt className="text-text-muted">{label}</dt>
      <dd className="truncate text-text-main">{value}</dd>
    </div>
  );
}
