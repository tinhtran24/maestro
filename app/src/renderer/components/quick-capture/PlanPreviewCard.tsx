import { Bot, FileText } from "lucide-react";
import type { TaskDraft } from "./types";

// PlanPreviewCard shows the planner's initial high-level execution plan and the
// files it expects to touch (read-only preview).
export function PlanPreviewCard({ draft }: { draft: TaskDraft }) {
	return (
		<div className="space-y-3">
			<div className="flex items-center gap-2 text-[13px]">
				<Bot className="size-4 text-violet-400" aria-hidden="true" />
				<span className="font-medium text-foreground">AI Plan Preview</span>
			</div>
			<p className="text-[12px] text-muted-foreground">The planner will use this context to create an execution plan.</p>

			{draft.plan?.length ? (
				<div className="space-y-1.5">
					<span className="text-[11px] font-medium text-muted-foreground">Initial plan (high level)</span>
					<ol className="space-y-1">
						{draft.plan.map((step, i) => (
							<li key={`${i}-${step}`} className="flex gap-2 text-[12px] text-foreground">
								<span className="shrink-0 tabular-nums text-violet-400">{i + 1}.</span>
								<span>{step}</span>
							</li>
						))}
					</ol>
				</div>
			) : null}

			{draft.likelyFiles?.length ? (
				<div className="space-y-1.5">
					<span className="text-[11px] font-medium text-muted-foreground">Likely files</span>
					<div className="flex flex-wrap gap-1.5">
						{draft.likelyFiles.map((f) => (
							<span key={f} className="inline-flex items-center gap-1 rounded-md border border-border bg-surface px-1.5 py-0.5 font-mono text-[11px] text-muted-foreground">
								<FileText className="size-3" aria-hidden="true" />
								{f}
							</span>
						))}
					</div>
				</div>
			) : null}
		</div>
	);
}
