import { describe, expect, it } from "vitest";
import { suggestTaskMetadata, taskCreationBody } from "./CreateTaskWizard";

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
		});
		expect(body).not.toHaveProperty("model");
	});

	it("suggests branch and finalization metadata when the task is created", () => {
		expect(suggestTaskMetadata("Fix issue: tracker intake credential diagnostics")).toEqual({
			branch: "bugfix/tracker-intake-credential-diagnostics",
			commitMessage: "fix(tracker): tracker intake credential diagnostics",
			prTitle: "fix(tracker): tracker intake credential diagnostics",
		});
	});
});
