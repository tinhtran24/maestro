import { apiClient, apiErrorMessage } from "./api-client";
import type { TaskDraft, Attachment } from "../components/quick-capture/types";

// extractTask sends raw Quick Capture input to the daemon's planner
// (POST /api/v1/plan), which runs the chosen agent CLI headlessly and returns a
// structured task draft. `agent` selects the planner agent (global/project
// default resolved by the caller).
export async function extractTask(params: {
	input: string;
	attachments?: Attachment[];
	agent?: string;
	projectId?: string;
}): Promise<{ draft: TaskDraft; agent: string }> {
	const attachmentDescriptions = (params.attachments ?? []).map((a) =>
		a.url ? `${a.kind}: ${a.name} (${a.url})` : `${a.kind}: ${a.name}`,
	);

	const { data, error } = await apiClient.POST("/api/v1/plan", {
		body: {
			input: params.input,
			attachments: attachmentDescriptions,
			agent: params.agent || undefined,
			projectId: params.projectId || undefined,
		},
	});
	if (error) throw new Error(apiErrorMessage(error, "AI could not structure the task"));
	if (!data?.draft) throw new Error("Planner returned no draft");
	return { draft: data.draft, agent: data.agent };
}

// composeSessionPrompt flattens the reviewed draft into the worker prompt that
// POST /api/v1/sessions receives — so the agent starts with the full structured
// brief, not just a title.
export function composeSessionPrompt(draft: TaskDraft, extraContext?: string): string {
	const lines: string[] = [];
	if (draft.description) lines.push(draft.description.trim(), "");
	if (draft.userStory) lines.push(`User story: ${draft.userStory}`, "");
	if (draft.acceptanceCriteria?.length) {
		lines.push("Acceptance criteria:");
		for (const c of draft.acceptanceCriteria) lines.push(`- ${c}`);
		lines.push("");
	}
	if (draft.technicalNotes) lines.push("Technical notes:", draft.technicalNotes, "");
	if (draft.likelyFiles?.length) lines.push(`Likely files: ${draft.likelyFiles.join(", ")}`, "");
	if (draft.risks?.length) {
		lines.push("Risks:");
		for (const r of draft.risks) lines.push(`- ${r}`);
		lines.push("");
	}
	if (draft.openQuestions?.length) {
		lines.push("Open questions:");
		for (const q of draft.openQuestions) lines.push(`- ${q}`);
		lines.push("");
	}
	if (extraContext?.trim()) lines.push("Additional context:", extraContext.trim(), "");
	return lines.join("\n").trim();
}
