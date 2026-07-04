import { FileText, Paperclip } from "lucide-react";
import type { CapturedAttachment, ExtractedTask } from "../../features/tasks/quickCapture";
import { ImagePreviewGrid } from "./ImagePreviewGrid";
import { MissingInformationCard } from "./MissingInformationCard";
import { PlanPreviewCard } from "./PlanPreviewCard";

// Step 3 — Review & Edit. Task details summary, attachments, and the AI plan
// preview. Attachments remain editable (remove/reorder); field edits happen in
// step 2, shown here read-only for a final look.
export function ReviewTaskFlow({
  draft,
  attachments,
  onRemove,
  onMove,
}: {
  draft: ExtractedTask;
  attachments: CapturedAttachment[];
  onRemove: (id: string) => void;
  onMove: (id: string, direction: -1 | 1) => void;
}) {
  return (
    <div className="grid gap-4">
      <section className="grid gap-2 rounded-lg border border-slate-800 bg-slate-950/40 p-3">
        <h4 className="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-text-muted"><FileText size={13} /> Task details</h4>
        <div className="flex items-center gap-2">
          <span className="rounded-md border border-purple-primary/40 bg-purple-primary/10 px-2 py-0.5 text-[11px] text-purple-hover">{draft.priority.value}</span>
          <h3 className="text-base font-semibold text-text-main">{draft.title.value}</h3>
        </div>
        <p className="text-sm text-text-muted">{draft.description.value}</p>
        <div className="flex flex-wrap gap-1.5">
          {draft.labels.value.map((label) => <span key={label} className="rounded-md border border-slate-700 bg-slate-800/60 px-2 py-0.5 text-[11px] text-text-muted">{label}</span>)}
        </div>
        <p className="text-xs text-text-muted">Feature: <span className="text-text-main">{draft.feature.value}</span> · Owner: <span className="text-text-main">{draft.assignedAgent}</span></p>
      </section>

      {attachments.length > 0 && (
        <section className="grid gap-2">
          <h4 className="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-text-muted"><Paperclip size={13} /> Attachments</h4>
          <ImagePreviewGrid attachments={attachments} onRemove={onRemove} onMove={onMove} />
        </section>
      )}

      <section className="grid gap-2">
        <h4 className="text-xs font-semibold uppercase tracking-wide text-text-muted">AI plan preview</h4>
        <PlanPreviewCard draft={draft} />
      </section>

      <MissingInformationCard questions={draft.openQuestions} />
    </div>
  );
}
