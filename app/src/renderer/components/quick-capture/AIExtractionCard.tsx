import { Sparkles } from "lucide-react";
import type { TaskDraft } from "./types";
import { ConfidenceBadge } from "./ConfidenceBadge";
import { Field, LabelChips, PrioritySelect, StringList, TextArea, TextField } from "./fields";

// AIExtractionCard is Step 2: the planner's auto-extracted, fully-editable task
// fields with per-field confidence badges.
export function AIExtractionCard({
	draft,
	onChange,
	agent,
}: {
	draft: TaskDraft;
	onChange: (patch: Partial<TaskDraft>) => void;
	agent?: string;
}) {
	return (
		<div className="space-y-4">
			<div className="flex items-center gap-2 rounded-lg border border-violet-500/25 bg-violet-500/5 px-3 py-2 text-[12px] text-violet-200">
				<Sparkles className="size-4 shrink-0" aria-hidden="true" />
				<span>
					{agent ? `${agent} extracted the following task` : "AI extracted the following task"} — review and edit before
					creating.
				</span>
			</div>

			<div className="grid gap-4 sm:grid-cols-[1fr_140px]">
				<Field label="Title" hint={<ConfidenceBadge value={draft.confidence?.title ?? 0} />}>
					<TextField value={draft.title} onChange={(v) => onChange({ title: v })} placeholder="Task title" />
				</Field>
				<Field label="Priority" hint={<ConfidenceBadge value={draft.confidence?.priority ?? 0} />}>
					<PrioritySelect value={draft.priority} onChange={(v) => onChange({ priority: v })} />
				</Field>
			</div>

			<Field label="Description">
				<TextArea value={draft.description} onChange={(v) => onChange({ description: v })} rows={3} placeholder="What needs to change and why" />
			</Field>

			<Field label="Acceptance criteria" hint={<ConfidenceBadge value={draft.confidence?.acceptanceCriteria ?? 0} />}>
				<StringList
					check
					values={draft.acceptanceCriteria ?? []}
					onChange={(v) => onChange({ acceptanceCriteria: v })}
					placeholder="Add criterion"
				/>
			</Field>

			<div className="grid gap-4 sm:grid-cols-2">
				<Field label="Labels">
					<LabelChips values={draft.labels ?? []} onChange={(v) => onChange({ labels: v })} />
				</Field>
				<Field label="Estimate">
					<TextField value={draft.estimate} onChange={(v) => onChange({ estimate: v })} placeholder="e.g. 4h" />
				</Field>
			</div>
		</div>
	);
}
