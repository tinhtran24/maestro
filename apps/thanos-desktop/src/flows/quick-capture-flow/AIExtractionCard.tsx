import { Bot } from "lucide-react";
import type { Priority } from "../../domain/models";
import type { ExtractedTask } from "../../features/tasks/quickCapture";
import { ConfidenceBadge } from "./ConfidenceBadge";

// Step 2 — AI Structure. The mock planner's inferred fields, each editable, with a
// confidence indicator. Users can modify everything before review.
export function AIExtractionCard({
  draft,
  onTitle,
  onDescription,
  onPriority,
  onLabels,
  onAcceptance,
  onTechnicalNotes,
}: {
  draft: ExtractedTask;
  onTitle: (value: string) => void;
  onDescription: (value: string) => void;
  onPriority: (value: Priority) => void;
  onLabels: (value: string[]) => void;
  onAcceptance: (value: string[]) => void;
  onTechnicalNotes: (value: string) => void;
}) {
  return (
    <div className="grid gap-4">
      <p className="inline-flex items-center gap-2 text-sm text-text-muted">
        <Bot size={15} className="text-purple-hover" /> AI structured your input into an engineering task. Edit anything.
      </p>

      <Row label="Title" confidence={draft.title.confidence}>
        <input value={draft.title.value} onChange={(event) => onTitle(event.target.value)} className="h-9 w-full rounded-lg border border-slate-800 bg-slate-950/60 px-3 text-sm text-text-main focus:border-slate-600 focus:outline-none" />
      </Row>

      <Row label="Description" confidence={draft.description.confidence}>
        <textarea value={draft.description.value} onChange={(event) => onDescription(event.target.value)} rows={2} className="w-full resize-y rounded-lg border border-slate-800 bg-slate-950/60 px-3 py-2 text-sm text-text-main focus:border-slate-600 focus:outline-none" />
      </Row>

      <div className="grid gap-4 sm:grid-cols-2">
        <Row label="Priority" confidence={draft.priority.confidence}>
          <select value={draft.priority.value} onChange={(event) => onPriority(event.target.value as Priority)} className="h-9 w-full rounded-lg border border-slate-800 bg-slate-950/60 px-3 text-sm text-text-main focus:border-slate-600 focus:outline-none">
            {(["P0", "P1", "P2", "P3"] satisfies Priority[]).map((priority) => <option key={priority}>{priority}</option>)}
          </select>
        </Row>
        <Row label="Feature" confidence={draft.feature.confidence}>
          <div className="flex h-9 items-center rounded-lg border border-slate-800 bg-slate-950/40 px-3 text-sm text-text-muted">{draft.feature.value}</div>
        </Row>
      </div>

      <Row label="Labels" confidence={draft.labels.confidence}>
        <input
          value={draft.labels.value.join(", ")}
          onChange={(event) => onLabels(event.target.value.split(",").map((item) => item.trim()).filter(Boolean))}
          placeholder="comma, separated, labels"
          className="h-9 w-full rounded-lg border border-slate-800 bg-slate-950/60 px-3 text-sm text-text-main focus:border-slate-600 focus:outline-none"
        />
      </Row>

      <div>
        <h4 className="mb-1 text-xs font-semibold uppercase tracking-wide text-text-muted">Acceptance Criteria</h4>
        <textarea
          value={draft.acceptanceCriteria.join("\n")}
          onChange={(event) => onAcceptance(event.target.value.split("\n").map((item) => item.trim()).filter(Boolean))}
          rows={4}
          placeholder="One criterion per line"
          className="w-full resize-y rounded-lg border border-slate-800 bg-slate-950/60 px-3 py-2 text-sm text-text-main focus:border-slate-600 focus:outline-none"
        />
      </div>

      <div>
        <h4 className="mb-1 text-xs font-semibold uppercase tracking-wide text-text-muted">Technical Notes</h4>
        <textarea value={draft.technicalNotes} onChange={(event) => onTechnicalNotes(event.target.value)} rows={2} className="w-full resize-y rounded-lg border border-slate-800 bg-slate-950/60 px-3 py-2 text-sm text-text-main focus:border-slate-600 focus:outline-none" />
      </div>
    </div>
  );
}

function Row({ label, confidence, children }: { label: string; confidence: number; children: React.ReactNode }) {
  return (
    <label className="grid gap-1">
      <span className="flex items-center justify-between gap-2 text-xs font-semibold uppercase tracking-wide text-text-muted">
        {label} <ConfidenceBadge value={confidence} />
      </span>
      {children}
    </label>
  );
}
