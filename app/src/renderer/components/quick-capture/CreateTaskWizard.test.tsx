import { describe, expect, it } from "vitest";
import { TASK_COMPLETION_FOOTER, suggestTaskMetadata, taskCreationBody } from "./CreateTaskWizard";

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

	it("suggests branch and PR metadata when the task is created", () => {
		expect(suggestTaskMetadata("Fix issue: tracker intake credential diagnostics")).toEqual({
			branch: "bugfix/tracker-intake-credential-diagnostics",
			commitMessage: "fix(tracker): tracker intake credential diagnostics",
			prTitle: "fix(tracker): tracker intake credential diagnostics",
		});
	});
});
