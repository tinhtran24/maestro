import { describe, expect, it } from "vitest";
import { TASK_COMPLETION_FOOTER, reviewedTaskMetadata, taskCreationBody } from "./CreateTaskWizard";
import { emptyDraft } from "./types";

describe("taskCreationBody", () => {
	it("omits native CLI model overrides while preserving the selected execution agent", () => {
		const body = taskCreationBody({
			projectId: "proj-1",
			agent: "codex",
			agentTouched: true,
			issueId: "Implement task creation",
			prompt: "Implement task creation",
			suggestions: {
				branch: "feature/implement-task-creation",
				commitMessage: "feat: implement task creation",
				prTitle: "feat: implement task creation",
			},
		});

		expect(body).toMatchObject({
			projectId: "proj-1",
			kind: "worker",
			harness: "codex",
			branch: "feature/implement-task-creation",
			commitMessage: "feat: implement task creation",
			prTitle: "feat: implement task creation",
			prompt: `Implement task creation\n\n${TASK_COMPLETION_FOOTER}`,
		});
		expect(body).not.toHaveProperty("model");
	});

	it("does not duplicate the completion footer when the prompt already includes it", () => {
		const body = taskCreationBody({
			projectId: "proj-1",
			agent: "codex",
			agentTouched: true,
			issueId: "Implement task creation",
			prompt: `Implement task creation\n\n${TASK_COMPLETION_FOOTER}`,
		});

		expect(body.prompt).toBe(`Implement task creation\n\n${TASK_COMPLETION_FOOTER}`);
	});

	it("uses the metadata edited during Review & Edit when the task is created", () => {
		expect(
			reviewedTaskMetadata({
				...emptyDraft(),
				title: "A title that would generate different metadata",
				suggestedBranch: "bugfix/custom-reviewed-branch",
				suggestedCommit: "fix(planner): preserve reviewed metadata",
			}),
		).toEqual({
			branch: "bugfix/custom-reviewed-branch",
			commitMessage: "fix(planner): preserve reviewed metadata",
			prTitle: "fix(planner): preserve reviewed metadata",
		});
	});
});
