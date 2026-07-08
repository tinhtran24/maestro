import { HelpCircle } from "lucide-react";
import type { TaskDraft } from "./types";

// MissingInformationCard surfaces the facts the planner still needs. These are
// not blockers — the planner agent can ask them later during execution.
export function MissingInformationCard({ draft }: { draft: TaskDraft }) {
	const items = [...(draft.missingInformation ?? []), ...(draft.openQuestions ?? [])];
	if (items.length === 0) return null;
	return (
		<div className="space-y-2 rounded-lg border border-amber-500/25 bg-amber-500/5 p-3">
			<div className="flex items-center gap-2 text-[12px] font-medium text-amber-200">
				<HelpCircle className="size-4" aria-hidden="true" />
				Missing information
			</div>
			<div className="flex flex-wrap gap-1.5">
				{items.map((q, i) => (
					<span key={`${i}-${q}`} className="rounded-full border border-amber-500/30 bg-amber-500/10 px-2 py-0.5 text-[11px] text-amber-100">
						{q}
					</span>
				))}
			</div>
			<p className="text-[11px] text-amber-200/70">The planner can ask these questions later — you don’t need to answer now.</p>
		</div>
	);
}
