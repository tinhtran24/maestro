import { useState } from "react";
import type { Attachment, TaskDraft } from "./types";
import { cn } from "../../lib/utils";
import { Field, LabelChips, PrioritySelect, StringList, TextArea, TextField } from "./fields";
import { ImagePreviewGrid } from "./ImagePreviewGrid";
import { PlanPreviewCard } from "./PlanPreviewCard";
import { MissingInformationCard } from "./MissingInformationCard";

type ReviewTab = "details" | "attachments" | "plan";

// TaskReviewFlow is Step 3: Details / Attachments / AI Plan Preview tabs over
// the same editable draft, plus an optional free-text context box.
export function TaskReviewFlow({
	draft,
	onChange,
	attachments,
	onRemove,
	onReorder,
	extraContext,
	onExtraContext,
}: {
	draft: TaskDraft;
	onChange: (patch: Partial<TaskDraft>) => void;
	attachments: Attachment[];
	onRemove: (id: string) => void;
	onReorder: (id: string, dir: -1 | 1) => void;
	extraContext: string;
	onExtraContext: (v: string) => void;
}) {
	const [tab, setTab] = useState<ReviewTab>("details");
	const tabs: { id: ReviewTab; label: string }[] = [
		{ id: "details", label: "Details" },
		{ id: "attachments", label: `Attachments${attachments.length ? ` (${attachments.length})` : ""}` },
		{ id: "plan", label: "AI Plan Preview" },
	];

	return (
		<div className="space-y-4">
			<div className="flex gap-1 border-b border-border">
				{tabs.map((t) => (
					<button
						key={t.id}
						type="button"
						onClick={() => setTab(t.id)}
						className={cn(
							"-mb-px border-b-2 px-3 py-1.5 text-[12px] transition",
							tab === t.id
								? "border-violet-500 text-foreground"
								: "border-transparent text-muted-foreground hover:text-foreground",
						)}
					>
						{t.label}
					</button>
				))}
			</div>

			{tab === "details" ? (
				<div className="space-y-4">
					<Field label="Title">
						<TextField value={draft.title} onChange={(v) => onChange({ title: v })} />
					</Field>
					<Field label="User story">
						<TextArea
							value={draft.userStory}
							onChange={(v) => onChange({ userStory: v })}
							rows={2}
							placeholder="As a … I want … so that …"
						/>
					</Field>
					<Field label="Analysis">
						<TextArea
							value={draft.analysis}
							onChange={(v) => onChange({ analysis: v })}
							rows={3}
							placeholder="Problem understanding and approach"
						/>
					</Field>
					<Field label="Priority">
						<PrioritySelect value={draft.priority} onChange={(v) => onChange({ priority: v })} />
					</Field>
					<Field label="Labels">
						<LabelChips values={draft.labels ?? []} onChange={(v) => onChange({ labels: v })} />
					</Field>
					<Field label="Technical notes">
						<TextArea value={draft.technicalNotes} onChange={(v) => onChange({ technicalNotes: v })} rows={4} />
					</Field>
					<Field label="Likely files">
						<StringList
							values={draft.likelyFiles ?? []}
							onChange={(v) => onChange({ likelyFiles: v })}
							placeholder="src/…"
						/>
					</Field>
					<Field label="Suggested branch">
						<TextField
							value={draft.suggestedBranch}
							onChange={(v) => onChange({ suggestedBranch: v })}
							placeholder="feature/short-description"
						/>
					</Field>
					<Field label="Suggested commit">
						<TextField
							value={draft.suggestedCommit}
							onChange={(v) => onChange({ suggestedCommit: v })}
							placeholder="feat(scope): summary"
						/>
					</Field>
					<Field label="Additional context (optional)">
						<TextArea
							value={extraContext}
							onChange={onExtraContext}
							rows={2}
							placeholder="Anything else the agent should know"
						/>
					</Field>
				</div>
			) : null}

			{tab === "attachments" ? (
				attachments.length ? (
					<ImagePreviewGrid attachments={attachments} onRemove={onRemove} onReorder={onReorder} />
				) : (
					<p className="py-6 text-center text-[12px] text-muted-foreground">No attachments.</p>
				)
			) : null}

			{tab === "plan" ? (
				<div className="space-y-4">
					<PlanPreviewCard draft={draft} />
					<MissingInformationCard draft={draft} />
				</div>
			) : null}
		</div>
	);
}
