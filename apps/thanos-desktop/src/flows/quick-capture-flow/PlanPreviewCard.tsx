import { AlertTriangle, FileCode2, ListChecks, ScrollText } from "lucide-react";
import type { ExtractedTask } from "../../features/tasks/quickCapture";

// AI Plan Preview — the planner's summary, likely files, and risks derived from the
// captured input. Read-only preview (the real plan is generated later in Planning).
export function PlanPreviewCard({ draft }: { draft: ExtractedTask }) {
  return (
    <div className="grid gap-3 rounded-lg border border-slate-800 bg-slate-950/40 p-3">
      <p className="inline-flex items-center gap-2 text-sm text-text-main">
        <ScrollText size={15} className="text-purple-hover" /> {draft.planSummary}
      </p>
      <Section icon={ListChecks} title="Acceptance criteria">
        <ul className="grid gap-1 text-xs text-text-muted">
          {draft.acceptanceCriteria.map((item) => <li key={item}>• {item}</li>)}
        </ul>
      </Section>
      <div className="grid gap-3 sm:grid-cols-2">
        <Section icon={FileCode2} title="Likely files">
          <ul className="grid gap-1 text-xs text-text-muted">
            {draft.likelyFiles.map((file) => <li key={file} className="font-mono">{file}</li>)}
          </ul>
        </Section>
        <Section icon={AlertTriangle} title="Risks">
          <ul className="grid gap-1 text-xs text-text-muted">
            {draft.risks.map((risk) => <li key={risk}>• {risk}</li>)}
          </ul>
        </Section>
      </div>
      <p className="text-xs text-text-muted">Estimated scope: <span className="text-text-main">{draft.estimatedScope}</span></p>
    </div>
  );
}

function Section({ icon: Icon, title, children }: { icon: typeof ListChecks; title: string; children: React.ReactNode }) {
  return (
    <div>
      <h4 className="mb-1 inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-text-muted"><Icon size={12} /> {title}</h4>
      {children}
    </div>
  );
}
