import type { components } from "../../../api/schema";

// The AI-structured task draft returned by POST /api/v1/plan. Every field is
// editable in the wizard before the task is created.
export type TaskDraft = components["schemas"]["PlannerTaskDraft"];
export type Confidence = components["schemas"]["PlannerConfidence"];

// A pasted / dropped / uploaded attachment. Images carry a data URL for preview;
// links (Figma/GitHub/Jira) carry a URL. Only name/type are sent to the planner.
export type AttachmentKind = "image" | "pdf" | "markdown" | "text" | "figma" | "github" | "jira" | "url" | "file";

export type Attachment = {
	id: string;
	kind: AttachmentKind;
	name: string;
	/** Object/data URL for image previews. */
	previewUrl?: string;
	/** External URL for link attachments. */
	url?: string;
	/** Human size label, e.g. "128 KB". */
	sizeLabel?: string;
};

// The four wizard steps.
export type WizardStep = "capture" | "structure" | "review" | "create";

// An empty draft the wizard falls back to before/while extracting.
export function emptyDraft(): TaskDraft {
	return {
		title: "",
		description: "",
		userStory: "",
		priority: "P2",
		labels: [],
		acceptanceCriteria: [],
		technicalNotes: "",
		likelyFiles: [],
		risks: [],
		dependencies: [],
		estimate: "",
		scope: "",
		plan: [],
		openQuestions: [],
		missingInformation: [],
		suggestedAgent: "",
		confidence: { overall: 0, title: 0, priority: 0, acceptanceCriteria: 0 },
	};
}

// classifyUrl maps a pasted URL to a link attachment kind.
export function classifyUrl(url: string): AttachmentKind {
	const u = url.toLowerCase();
	if (u.includes("figma.com")) return "figma";
	if (u.includes("github.com")) return "github";
	if (u.includes("atlassian.net") || u.includes("/jira")) return "jira";
	return "url";
}
